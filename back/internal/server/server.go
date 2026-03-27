package server

import (
	"context"
	"log/slog"
	"net/http"
	"os"

	"github.com/tanguyRa/saas_seed/internal/config"
	"github.com/tanguyRa/saas_seed/internal/handlers"
	"github.com/tanguyRa/saas_seed/internal/repository"

	"github.com/jackc/pgx/v5/pgxpool"
	"github.com/lestrrat-go/jwx/v3/jwk"
)

// Server represents the HTTP server
type Server struct {
	config                 config.Config
	logger                 *slog.Logger
	pool                   *pgxpool.Pool
	queries                *repository.Queries
	handlers               *handlers.Handlers
	authVerificationKeyset *jwk.Set
}

// New creates a new Server with the given configuration
func New(cfg config.Config) *Server {
	logger := slog.New(slog.NewTextHandler(os.Stdout, nil))

	return &Server{
		config: cfg,
		logger: logger,
	}
}

// Start initializes the database connection, Docker client, and starts the HTTP server
func (s *Server) Start() error {
	ctx := context.Background()

	// Connect to database
	pool, err := pgxpool.New(ctx, s.config.Database.ConnectionString)
	if err != nil {
		s.logger.Error("failed to connect to database", "error", err)
		return err
	}
	defer pool.Close()
	s.pool = pool

	// Verify connection
	if err := pool.Ping(ctx); err != nil {
		s.logger.Error("failed to ping database", "error", err)
		return err
	}
	s.logger.Info("connected to database")

	// Create repository queries
	s.queries = repository.New(pool)

	// Create handlers (pass pool for transaction support)
	handlerSet, err := handlers.New(s.queries, s.logger, s.config)
	if err != nil {
		s.logger.Error("failed to initialize handlers", "error", err)
		return err
	}
	s.handlers = handlerSet

	// Setup routes
	handler := s.initRoutes()

	return http.ListenAndServe(":8080", handler)
}

// GetDockerClient returns the Docker client with read lock
func (s *Server) Shutdown() {
	s.pool.Close()
}
