package handlers

import (
	"context"

	"github.com/tanguyRa/saas_seed/internal/repository"
)

type queriesContextKey struct{}

func WithQueries(ctx context.Context, queries *repository.Queries) context.Context {
	return context.WithValue(ctx, queriesContextKey{}, queries)
}

func queriesFromContext(ctx context.Context, fallback *repository.Queries) *repository.Queries {
	if queries, ok := ctx.Value(queriesContextKey{}).(*repository.Queries); ok && queries != nil {
		return queries
	}
	return fallback
}
