package server

import (
	"context"
	"net/http"
	"net/http/httptest"
	"net/url"
	"os"
	"testing"

	"github.com/tanguyRa/onvaou/internal/session"

	"log/slog"

	"github.com/jackc/pgx/v5/pgxpool"
	"github.com/lestrrat-go/jwx/v3/jwa"
	"github.com/lestrrat-go/jwx/v3/jwt"
)

func TestAuthenticateMiddleware(t *testing.T) {
	s := &Server{logger: slog.New(slog.NewTextHandler(os.Stdout, nil))}

	token := jwt.New()
	_ = token.Set(jwt.SubjectKey, "user-123")
	_ = token.Set("email", "user@example.com")
	_ = token.Set("name", "User")
	signed, err := jwt.Sign(token, jwt.WithKey(jwa.HS256(), []byte("secret")))
	if err != nil {
		t.Fatalf("sign token: %v", err)
	}

	req := httptest.NewRequest(http.MethodGet, "/api/secured/ping", nil)
	req.Header.Set("Authorization", "Bearer "+string(signed))

	rec := httptest.NewRecorder()
	handler := s.authenticate(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		user, ok := session.UserFromContext(r.Context())
		if !ok || user == nil || user.ID != "user-123" {
			t.Fatalf("expected authenticated user in context")
		}
		w.WriteHeader(http.StatusOK)
	}))
	handler.ServeHTTP(rec, req)

	if rec.Code != http.StatusOK {
		t.Fatalf("expected 200, got %d", rec.Code)
	}
}

func TestRequireAuthenticationMiddleware(t *testing.T) {
	s := &Server{logger: slog.New(slog.NewTextHandler(os.Stdout, nil))}
	req := httptest.NewRequest(http.MethodGet, "/api/library", nil)

	rec := httptest.NewRecorder()
	handler := s.requireAuthentication(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusOK)
	}))
	handler.ServeHTTP(rec, req)

	if rec.Code != http.StatusUnauthorized {
		t.Fatalf("expected 401, got %d", rec.Code)
	}
}

func TestWithDBSessionMiddleware(t *testing.T) {
	pool := requireLocalDB(t)
	defer pool.Close()

	s := &Server{
		logger: slog.New(slog.NewTextHandler(os.Stdout, nil)),
		pool:   pool,
	}

	req := httptest.NewRequest(http.MethodGet, "/api/ping", nil)
	req.Header.Set("Authorization", "Bearer test-token")

	rec := httptest.NewRecorder()
	called := false
	handler := s.withDBSession(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		called = true
		w.WriteHeader(http.StatusOK)
	}))
	handler.ServeHTTP(rec, req)

	if !called {
		t.Fatalf("expected handler to be called")
	}
	if rec.Code != http.StatusOK {
		t.Fatalf("expected 200, got %d", rec.Code)
	}
}

func requireLocalDB(t *testing.T) *pgxpool.Pool {
	t.Helper()
	dsn := os.Getenv("DATABASE_URL")
	if dsn == "" {
		t.Skip("DATABASE_URL not set")
	}
	parsed, err := url.Parse(dsn)
	if err == nil {
		host := parsed.Hostname()
		if host != "" && host != "localhost" && host != "127.0.0.1" && host != "postgres" {
			t.Skip("DATABASE_URL is not local; set a local DATABASE_URL for integration tests")
		}
	}
	pool, err := pgxpool.New(context.Background(), dsn)
	if err != nil {
		t.Fatalf("failed to connect to db: %v", err)
	}
	return pool
}
