package handlers

import (
	"context"
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/tanguyRa/onvaou/internal/session"
)

func TestAuthUserFromRequest(t *testing.T) {
	handler := NewAuthHandler(nil, makeLogger())
	user := &session.UserInfo{ID: "user-1", Email: "user@example.com", Name: "User"}
	ctx := context.WithValue(context.Background(), session.UserContextKey, user)

	req := httptest.NewRequest(http.MethodGet, "/api/secured/ping", nil).WithContext(ctx)
	rec := httptest.NewRecorder()
	handler.UserFromRequest(rec, req)

	if rec.Code != http.StatusOK {
		t.Fatalf("expected 200, got %d", rec.Code)
	}
	requireContains(t, rec.Body.String(), "\"id\":\"user-1\"")
}
