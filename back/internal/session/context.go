package session

import (
	"context"
)

type contextKey string

const IsAuthenticatedContextKey = contextKey("isAuthenticated")
const UserContextKey = contextKey("userSession")

type UserInfo struct {
	ID    string `json:"id"`
	Email string `json:"email"`
	Name  string `json:"name"`
}

// Inside your auth package or a common utility file
func UserFromContext(ctx context.Context) (*UserInfo, bool) {
	user, ok := ctx.Value(UserContextKey).(*UserInfo)
	return user, ok
}
