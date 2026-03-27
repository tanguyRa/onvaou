package main

import (
	"log"
	"net/http"

	"github.com/tanguyRa/saas_seed/internal/config"
	"github.com/tanguyRa/saas_seed/internal/server"
)

func main() {
	cfg, err := config.Load()
	if err != nil {
		log.Fatalf("Failed to load config: %v", err)
	}

	srv := server.New(*cfg)
	if err != nil {
		log.Fatalf("Failed to create server: %v", err)
	}

	// Start server in goroutine
	log.Printf("Server starting on %s", cfg.Address)
	if err := srv.Start(); err != nil && err != http.ErrServerClosed {
		log.Fatalf("Server error: %v", err)
	}
}
