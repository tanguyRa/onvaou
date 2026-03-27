package llmclient

import (
	"context"
	"fmt"

	"github.com/tanguyRa/saas_seed/internal/config"
	"github.com/tanguyRa/saas_seed/internal/llm"
)

type Client struct {
	provider llm.Provider
}

func NewFromConfig(cfg config.Config) (*Client, error) {
	providerName := cfg.LLM.Provider
	if providerName == "" {
		providerName = "gemini"
	}

	apiKey, model, err := selectCredentials(providerName, cfg)
	if err != nil {
		return nil, err
	}

	provider, err := llm.NewProvider(providerName, apiKey, model)
	if err != nil {
		return nil, err
	}

	return &Client{provider: provider}, nil
}

func NewWithProvider(provider llm.Provider) *Client {
	return &Client{provider: provider}
}

func (c *Client) Name() string {
	return c.provider.Name()
}

func (c *Client) Generate(ctx context.Context, systemPrompt string, messages []llm.Message) (string, error) {
	return c.provider.Chat(ctx, systemPrompt, messages)
}

func selectCredentials(provider string, cfg config.Config) (string, string, error) {
	switch provider {
	case "google":
		return cfg.LLM.Google.APIKey, cfg.LLM.Google.Model, nil
	case "anthropic":
		return cfg.LLM.Anthropic.APIKey, cfg.LLM.Anthropic.Model, nil
	case "openai":
		return cfg.LLM.OpenAI.APIKey, cfg.LLM.OpenAI.Model, nil
	default:
		return "", "", fmt.Errorf("unsupported LLM provider: %s", provider)
	}
}
