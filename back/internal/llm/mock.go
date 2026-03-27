package llm

import (
	"context"
)

// MockProvider is a test double for LLM providers
type MockProvider struct {
	name           string
	ChatResponse   string
	ChatError      error
	StreamChunks   []string
	StreamError    error
	ChatCalls      []MockChatCall
	StreamCalls    []MockStreamCall
}

type MockChatCall struct {
	SystemPrompt string
	Messages     []Message
}

type MockStreamCall struct {
	SystemPrompt string
	Messages     []Message
}

// NewMockProvider creates a mock provider with configurable responses
func NewMockProvider() *MockProvider {
	return &MockProvider{
		name:         "mock",
		ChatResponse: "Mock response",
		StreamChunks: []string{"Hello", " ", "World"},
	}
}

func (m *MockProvider) Name() string {
	return m.name
}

func (m *MockProvider) Chat(ctx context.Context, systemPrompt string, messages []Message) (string, error) {
	m.ChatCalls = append(m.ChatCalls, MockChatCall{
		SystemPrompt: systemPrompt,
		Messages:     messages,
	})

	if m.ChatError != nil {
		return "", m.ChatError
	}
	return m.ChatResponse, nil
}

func (m *MockProvider) ChatStream(ctx context.Context, systemPrompt string, messages []Message) (<-chan StreamChunk, error) {
	m.StreamCalls = append(m.StreamCalls, MockStreamCall{
		SystemPrompt: systemPrompt,
		Messages:     messages,
	})

	if m.StreamError != nil {
		return nil, m.StreamError
	}

	ch := make(chan StreamChunk)
	go func() {
		defer close(ch)
		for _, chunk := range m.StreamChunks {
			select {
			case <-ctx.Done():
				ch <- StreamChunk{Error: ctx.Err(), Done: true}
				return
			case ch <- StreamChunk{Content: chunk}:
			}
		}
		ch <- StreamChunk{Done: true}
	}()

	return ch, nil
}

// Reset clears recorded calls
func (m *MockProvider) Reset() {
	m.ChatCalls = nil
	m.StreamCalls = nil
}
