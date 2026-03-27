package server

import (
	"context"
	"fmt"
	"net/http"
	"strings"

	"github.com/tanguyRa/onvaou/internal/handlers"
	"github.com/tanguyRa/onvaou/internal/repository"
	"github.com/tanguyRa/onvaou/internal/session"

	"github.com/google/uuid"
	"github.com/lestrrat-go/jwx/v3/jwt"
)

func (s *Server) logRequest(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		var (
			ip     = r.RemoteAddr
			proto  = r.Proto
			method = r.Method
			uri    = r.URL.RequestURI()
		)

		s.logger.Info("received request", "ip", ip, "proto", proto, "method", method, "uri", uri)

		next.ServeHTTP(w, r)
	})
}

func (s *Server) recoverPanic(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		defer func() {
			pv := recover()
			if pv != nil {
				w.Header().Set("Connection", "close")
				s.serverError(w, r, fmt.Errorf("%v", pv))
			}
		}()

		next.ServeHTTP(w, r)
	})
}

func (s *Server) requireAuthentication(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		isAuthenticated, ok := r.Context().Value(session.IsAuthenticatedContextKey).(bool)
		if !ok || !isAuthenticated {
			s.clientError(w, http.StatusUnauthorized)
			return
		}

		next.ServeHTTP(w, r)
	})
}

func (s *Server) withDBSession(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if s.pool == nil {
			s.serverError(w, r, fmt.Errorf("database not initialized"))
			return
		}

		conn, err := s.pool.Acquire(r.Context())
		if err != nil {
			s.serverError(w, r, fmt.Errorf("failed to acquire db connection: %w", err))
			return
		}
		defer conn.Release()

		ctx := r.Context()

		_, _ = conn.Exec(ctx, "RESET app.user_id")

		if user, ok := session.UserFromContext(ctx); ok && user != nil {
			if _, err := uuid.Parse(user.ID); err == nil {
				_, _ = conn.Exec(ctx, "select set_config('app.user_id', $1, false)", user.ID)
			}
		}

		queries := repository.New(conn)
		ctx = handlers.WithQueries(ctx, queries)

		next.ServeHTTP(w, r.WithContext(ctx))
	})
}

func (s *Server) authenticate(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		token, err := jwt.ParseRequest(r, jwt.WithVerify(false))
		if err != nil {
			ctx := context.WithValue(r.Context(), session.IsAuthenticatedContextKey, false)
			next.ServeHTTP(w, r.WithContext(ctx))
			return
		}

		userID, exists := token.Subject()
		if !exists {
			ctx := context.WithValue(r.Context(), session.IsAuthenticatedContextKey, false)
			next.ServeHTTP(w, r.WithContext(ctx))
			return
		}

		var email string
		var name string

		token.Get("email", &email)
		token.Get("name", &name)

		userInfo := session.UserInfo{
			ID:    userID,
			Email: email,
			Name:  name,
		}

		// Create a new context with the user info
		ctx := context.WithValue(r.Context(), session.UserContextKey, &userInfo)
		ctx = context.WithValue(ctx, session.IsAuthenticatedContextKey, true)

		// Call the next handler with the new context
		next.ServeHTTP(w, r.WithContext(ctx))
	})
}

func isInternalRequest(r *http.Request, token string) bool {
	if strings.HasPrefix(r.URL.Path, "/webhooks/") {
		return true
	}
	if token == "" {
		return false
	}
	return r.Header.Get("Authorization") == "Bearer "+token
}

// requireUserOwnership validates that the userId path parameter matches the authenticated user
func (s *Server) requireUserOwnership(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		user, ok := session.UserFromContext(r.Context())
		if !ok || user == nil {
			s.clientError(w, http.StatusUnauthorized)
			return
		}

		pathUserID := r.PathValue("userId")
		if pathUserID != "" && pathUserID != user.ID {
			s.logger.Warn("user attempted to access another user's resource",
				"requestingUser", user.ID,
				"targetUser", pathUserID,
				"path", r.URL.Path,
			)
			s.clientError(w, http.StatusForbidden)
			return
		}

		next.ServeHTTP(w, r)
	})
}
