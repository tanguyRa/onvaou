package handlers

import (
	"encoding/json"
	"errors"
	"log/slog"
	"net/http"

	"github.com/tanguyRa/onvaou/internal/repository"
	"github.com/tanguyRa/onvaou/internal/session"
)

type AuthHandler struct {
	logger  *slog.Logger
	queries *repository.Queries
}

func NewAuthHandler(queries *repository.Queries, logger *slog.Logger) *AuthHandler {
	return &AuthHandler{
		queries: queries,
		logger:  logger,
	}
}

var (
	ErrMissingUserID = errors.New("missing user id")
)

func (h *AuthHandler) UserFromRequest(w http.ResponseWriter, r *http.Request) {
	h.logger.Debug("UserFromRequest called")
	userPtr, ok := session.UserFromContext(r.Context())
	h.logger.Debug("UserFromRequest", "userPtr", userPtr, "ok", ok)
	if !ok || userPtr == nil {
		http.Error(w, "Failed to retrieve user from context", http.StatusInternalServerError)
		return
	}

	// Marshal the user object to JSON
	jsonData, err := json.Marshal(userPtr) // Marshal the pointer directly
	if err != nil {
		h.logger.Error("Failed to marshal user data to JSON", "error", err) // Log the error
		http.Error(w, "Failed to encode user data", http.StatusInternalServerError)
		return
	}

	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(http.StatusOK)
	w.Write(jsonData)
}
