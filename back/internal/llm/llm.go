package llm

import (
	"context"
	"fmt"
)

// Message represents a chat message
type Message struct {
	Role    string `json:"role"` // "user" or "assistant"
	Content string `json:"content"`
}

// StreamChunk represents a chunk of streamed response
type StreamChunk struct {
	Content string `json:"content"`
	Done    bool   `json:"done"`
	Error   error  `json:"-"`
}

// Provider defines the interface for LLM providers
type Provider interface {
	// Chat sends a message and returns the full response
	Chat(ctx context.Context, systemPrompt string, messages []Message) (string, error)

	// ChatStream sends a message and streams the response
	ChatStream(ctx context.Context, systemPrompt string, messages []Message) (<-chan StreamChunk, error)

	// Name returns the provider name
	Name() string
}

// Config holds provider configuration
type Config struct {
	APIKey string
	Model  string
}

// NewProvider creates a provider based on the adapter name
func NewProvider(adapter, apiKey, model string) (Provider, error) {
	switch adapter {
	case "anthropic":
		return newAnthropicProvider(apiKey, model), nil
	case "google":
		return newGeminiProvider(apiKey, model), nil
	default:
		return newGeminiProvider(apiKey, model), fmt.Errorf("unknown adapter: %s", adapter)
	}
}
