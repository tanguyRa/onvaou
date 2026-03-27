package llm

import (
	"context"
	"errors"
	"strings"
	"testing"
	"time"
)

func TestMockProvider_Name(t *testing.T) {
	mock := NewMockProvider()
	if mock.Name() != "mock" {
		t.Errorf("expected name 'mock', got '%s'", mock.Name())
	}
}

func TestMockProvider_Chat(t *testing.T) {
	mock := NewMockProvider()
	mock.ChatResponse = "Test response"

	messages := []Message{
		{Role: "user", Content: "Hello"},
	}

	response, err := mock.Chat(context.Background(), "system prompt", messages)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	if response != "Test response" {
		t.Errorf("expected 'Test response', got '%s'", response)
	}

	if len(mock.ChatCalls) != 1 {
		t.Fatalf("expected 1 call, got %d", len(mock.ChatCalls))
	}

	if mock.ChatCalls[0].SystemPrompt != "system prompt" {
		t.Errorf("expected system prompt 'system prompt', got '%s'", mock.ChatCalls[0].SystemPrompt)
	}

	if len(mock.ChatCalls[0].Messages) != 1 || mock.ChatCalls[0].Messages[0].Content != "Hello" {
		t.Error("messages not recorded correctly")
	}
}

func TestMockProvider_Chat_Error(t *testing.T) {
	mock := NewMockProvider()
	mock.ChatError = errors.New("api error")

	_, err := mock.Chat(context.Background(), "", nil)
	if err == nil {
		t.Fatal("expected error, got nil")
	}

	if err.Error() != "api error" {
		t.Errorf("expected 'api error', got '%s'", err.Error())
	}
}

func TestMockProvider_ChatStream(t *testing.T) {
	mock := NewMockProvider()
	mock.StreamChunks = []string{"Hello", " ", "World", "!"}

	stream, err := mock.ChatStream(context.Background(), "system", []Message{{Role: "user", Content: "Hi"}})
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	var result strings.Builder
	for chunk := range stream {
		if chunk.Error != nil {
			t.Fatalf("unexpected stream error: %v", chunk.Error)
		}
		result.WriteString(chunk.Content)
		if chunk.Done {
			break
		}
	}

	if result.String() != "Hello World!" {
		t.Errorf("expected 'Hello World!', got '%s'", result.String())
	}

	if len(mock.StreamCalls) != 1 {
		t.Errorf("expected 1 stream call, got %d", len(mock.StreamCalls))
	}
}

func TestMockProvider_ChatStream_Error(t *testing.T) {
	mock := NewMockProvider()
	mock.StreamError = errors.New("stream init error")

	_, err := mock.ChatStream(context.Background(), "", nil)
	if err == nil {
		t.Fatal("expected error, got nil")
	}
}

func TestMockProvider_ChatStream_ContextCancellation(t *testing.T) {
	mock := NewMockProvider()
	mock.StreamChunks = []string{"1", "2", "3", "4", "5"}

	ctx, cancel := context.WithCancel(context.Background())

	stream, err := mock.ChatStream(ctx, "system", nil)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	// Read one chunk then cancel
	<-stream
	cancel()

	// Give goroutine time to process cancellation
	time.Sleep(10 * time.Millisecond)

	// Drain remaining
	for chunk := range stream {
		if chunk.Error == context.Canceled {
			return // Expected
		}
	}
}

func TestMockProvider_Reset(t *testing.T) {
	mock := NewMockProvider()

	mock.Chat(context.Background(), "", nil)
	mock.ChatStream(context.Background(), "", nil)

	if len(mock.ChatCalls) != 1 || len(mock.StreamCalls) != 1 {
		t.Fatal("calls should be recorded")
	}

	mock.Reset()

	if len(mock.ChatCalls) != 0 || len(mock.StreamCalls) != 0 {
		t.Error("Reset should clear all calls")
	}
}

func TestMessage_Structure(t *testing.T) {
	msg := Message{
		Role:    "assistant",
		Content: "Hello, how can I help?",
	}

	if msg.Role != "assistant" {
		t.Errorf("expected role 'assistant', got '%s'", msg.Role)
	}

	if msg.Content != "Hello, how can I help?" {
		t.Errorf("content mismatch")
	}
}

func TestStreamChunk_Structure(t *testing.T) {
	chunk := StreamChunk{
		Content: "test",
		Done:    false,
		Error:   nil,
	}

	if chunk.Content != "test" {
		t.Error("content mismatch")
	}

	if chunk.Done {
		t.Error("done should be false")
	}

	if chunk.Error != nil {
		t.Error("error should be nil")
	}
}
