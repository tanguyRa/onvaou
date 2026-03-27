package llm

import (
	"context"
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"
	"time"
)

func TestAnthropicProvider_Name(t *testing.T) {
	provider := newAnthropicProvider("test-key", "claude-sonnet-4-20250514")
	if provider.Name() != "anthropic" {
		t.Errorf("expected 'anthropic', got '%s'", provider.Name())
	}
}

func TestAnthropicProvider_DefaultModel(t *testing.T) {
	provider := newAnthropicProvider("test-key", "")
	if provider.model != "claude-sonnet-4-20250514" {
		t.Errorf("expected default model, got '%s'", provider.model)
	}
}

func TestAnthropicProvider_Chat_Success(t *testing.T) {
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		// Verify headers
		if r.Header.Get("x-api-key") != "test-api-key" {
			t.Errorf("expected api key header")
		}
		if r.Header.Get("anthropic-version") != "2023-06-01" {
			t.Errorf("expected anthropic-version header")
		}
		if r.Header.Get("Content-Type") != "application/json" {
			t.Errorf("expected content-type header")
		}

		// Verify request body
		var req anthropicRequest
		if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
			t.Fatalf("failed to decode request: %v", err)
		}

		if req.System != "You are helpful" {
			t.Errorf("expected system prompt, got '%s'", req.System)
		}
		if len(req.Messages) != 1 || req.Messages[0].Content != "Hello" {
			t.Errorf("messages not correct")
		}
		if req.Stream {
			t.Error("stream should be false for Chat")
		}

		// Return mock response
		w.Header().Set("Content-Type", "application/json")
		json.NewEncoder(w).Encode(anthropicResponse{
			Content: []struct {
				Type string `json:"type"`
				Text string `json:"text"`
			}{
				{Type: "text", Text: "Hello! How can I help you today?"},
			},
			StopReason: "end_turn",
		})
	}))
	defer server.Close()

// We need to make the provider use our test server via custom transport
	testProvider := &anthropicProvider{
		apiKey: "test-api-key",
		model:  "claude-sonnet-4-20250514",
		client: &http.Client{
			Transport: &testTransport{url: server.URL},
		},
	}

	response, err := testProvider.Chat(context.Background(), "You are helpful", []Message{
		{Role: "user", Content: "Hello"},
	})

	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	if response != "Hello! How can I help you today?" {
		t.Errorf("unexpected response: %s", response)
	}
}

func TestAnthropicProvider_Chat_APIError(t *testing.T) {
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusUnauthorized)
		w.Write([]byte(`{"error": {"message": "Invalid API key"}}`))
	}))
	defer server.Close()

	provider := &anthropicProvider{
		apiKey: "bad-key",
		model:  "claude-sonnet-4-20250514",
		client: &http.Client{
			Transport: &testTransport{url: server.URL},
		},
	}

	_, err := provider.Chat(context.Background(), "", []Message{})
	if err == nil {
		t.Fatal("expected error for unauthorized request")
	}

	if !strings.Contains(err.Error(), "401") {
		t.Errorf("error should contain status code: %v", err)
	}
}

func TestAnthropicProvider_Chat_EmptyResponse(t *testing.T) {
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Content-Type", "application/json")
		// Return response with empty content array
		w.Write([]byte(`{"content":[],"stop_reason":"end_turn"}`))
	}))
	defer server.Close()

	provider := &anthropicProvider{
		apiKey: "test-key",
		model:  "claude-sonnet-4-20250514",
		client: &http.Client{
			Transport: &testTransport{url: server.URL},
		},
	}

	response, err := provider.Chat(context.Background(), "", []Message{})
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	if response != "" {
		t.Errorf("expected empty response, got '%s'", response)
	}
}

func TestAnthropicProvider_ChatStream_Success(t *testing.T) {
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		// Verify stream is requested
		var req anthropicRequest
		json.NewDecoder(r.Body).Decode(&req)
		if !req.Stream {
			t.Error("stream should be true for ChatStream")
		}

		w.Header().Set("Content-Type", "text/event-stream")
		flusher := w.(http.Flusher)

		// Send streaming events
		events := []string{
			`data: {"type":"content_block_delta","index":0,"delta":{"type":"text_delta","text":"Hello"}}`,
			`data: {"type":"content_block_delta","index":0,"delta":{"type":"text_delta","text":" World"}}`,
			`data: {"type":"message_stop"}`,
		}

		for _, event := range events {
			w.Write([]byte(event + "\n"))
			flusher.Flush()
		}
	}))
	defer server.Close()

	provider := &anthropicProvider{
		apiKey: "test-key",
		model:  "claude-sonnet-4-20250514",
		client: &http.Client{
			Transport: &testTransport{url: server.URL},
		},
	}

	stream, err := provider.ChatStream(context.Background(), "system", []Message{
		{Role: "user", Content: "Hi"},
	})
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	var result strings.Builder
	for chunk := range stream {
		if chunk.Error != nil {
			t.Fatalf("stream error: %v", chunk.Error)
		}
		result.WriteString(chunk.Content)
		if chunk.Done {
			break
		}
	}

	if result.String() != "Hello World" {
		t.Errorf("expected 'Hello World', got '%s'", result.String())
	}
}

func TestAnthropicProvider_ChatStream_APIError(t *testing.T) {
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusBadRequest)
		w.Write([]byte(`{"error": "bad request"}`))
	}))
	defer server.Close()

	provider := &anthropicProvider{
		apiKey: "test-key",
		model:  "claude-sonnet-4-20250514",
		client: &http.Client{
			Transport: &testTransport{url: server.URL},
		},
	}

	_, err := provider.ChatStream(context.Background(), "", nil)
	if err == nil {
		t.Fatal("expected error")
	}
}

func TestAnthropicProvider_ChatStream_ContextCancellation(t *testing.T) {
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Content-Type", "text/event-stream")
		flusher := w.(http.Flusher)

		// Send chunks slowly
		for i := 0; i < 10; i++ {
			select {
			case <-r.Context().Done():
				return
			default:
				w.Write([]byte(`data: {"type":"content_block_delta","index":0,"delta":{"type":"text_delta","text":"chunk"}}` + "\n"))
				flusher.Flush()
				time.Sleep(50 * time.Millisecond)
			}
		}
	}))
	defer server.Close()

	provider := &anthropicProvider{
		apiKey: "test-key",
		model:  "claude-sonnet-4-20250514",
		client: &http.Client{
			Transport: &testTransport{url: server.URL},
		},
	}

	ctx, cancel := context.WithTimeout(context.Background(), 100*time.Millisecond)
	defer cancel()

	stream, err := provider.ChatStream(ctx, "", nil)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	// Read until context is cancelled
	count := 0
	for range stream {
		count++
	}

	// Should have received some chunks before cancellation
	if count == 0 {
		t.Error("expected at least one chunk")
	}
}

// testTransport redirects requests to the test server
type testTransport struct {
	url string
}

func (t *testTransport) RoundTrip(req *http.Request) (*http.Response, error) {
	// Redirect to test server
	req.URL.Scheme = "http"
	req.URL.Host = strings.TrimPrefix(t.url, "http://")
	return http.DefaultTransport.RoundTrip(req)
}
