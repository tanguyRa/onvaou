package server

import (
	"net/http"

	"github.com/tanguyRa/saas_seed/internal/middleware"
)

func (s *Server) initRoutes() http.Handler {
	mux := http.NewServeMux()

	mux.HandleFunc("GET /api/ping", s.handlers.Ping)

	dynamic := middleware.New(s.authenticate, s.withDBSession)
	// protected := dynamic.Append(s.requireAuthentication)

	mux.Handle("GET /api/secured/ping", dynamic.ThenFunc(s.handlers.Auth.UserFromRequest))
	// Webhooks
	mux.HandleFunc("POST /webhooks/polar", s.handlers.Polar.HandleWebhook)

	standard := middleware.New(s.recoverPanic, s.logRequest)
	return standard.Then(mux)
}
