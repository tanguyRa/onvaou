package handlers

import (
	"log/slog"
	"net/http"

	"github.com/tanguyRa/onvaou/internal/config"
	"github.com/tanguyRa/onvaou/internal/providers/llmclient"
	"github.com/tanguyRa/onvaou/internal/repository"
)

type Handlers struct {
	queries *repository.Queries
	logger  *slog.Logger
	config  config.Config
	llm     *llmclient.Client
	// store   *storage.
	Auth  *AuthHandler
	Polar *PolarHandler
}

// New creates a new Handlers instance
func New(queries *repository.Queries, logger *slog.Logger, cfg config.Config) (*Handlers, error) {
	llmClient, err := llmclient.NewFromConfig(cfg)
	if err != nil {
		return nil, err
	}

	// store, err := storage.NewMinIOStore(cfg)
	// if err != nil {
	// 	return nil, err
	// }

	h := &Handlers{
		queries: queries,
		logger:  logger,
		config:  cfg,
		llm:     llmClient,
		Auth:    NewAuthHandler(queries, logger),
		Polar:   NewPolarHandler(queries, logger, cfg.Polar),
	}
	return h, nil
}

func (h *Handlers) Ping(w http.ResponseWriter, r *http.Request) {
	w.Write([]byte("OK"))
}
