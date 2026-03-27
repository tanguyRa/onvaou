package handlers

import (
	"bytes"
	"context"
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"net/url"
	"os"
	"strings"
	"testing"
	"time"

	"github.com/tanguyRa/onvaou/internal/repository"
	"github.com/tanguyRa/onvaou/internal/session"

	"log/slog"

	"github.com/google/uuid"
	"github.com/jackc/pgx/v5/pgxpool"
)

type mockQueue struct {
	entries []queueEntry
	err     error
}

type queueEntry struct {
	jobType string
	payload any
}

func (m *mockQueue) Enqueue(ctx context.Context, jobType string, payload any) error {
	if m.err != nil {
		return m.err
	}
	m.entries = append(m.entries, queueEntry{jobType: jobType, payload: payload})
	return nil
}

type mockStore struct {
	data map[string][]byte
}

func newMockStore() *mockStore {
	return &mockStore{data: map[string][]byte{}}
}

func (m *mockStore) Put(ctx context.Context, key string, contentType string, data []byte) (string, error) {
	m.data[key] = append([]byte(nil), data...)
	return "mock://storage/" + key, nil
}

func (m *mockStore) Get(ctx context.Context, key string) ([]byte, error) {
	if data, ok := m.data[key]; ok {
		return append([]byte(nil), data...), nil
	}
	return nil, os.ErrNotExist
}

func requireLocalDB(t *testing.T) *pgxpool.Pool {
	t.Helper()
	dsn := os.Getenv("DATABASE_URL")
	if dsn == "" {
		t.Skip("DATABASE_URL not set")
	}
	parsed, err := url.Parse(dsn)
	if err == nil {
		host := parsed.Hostname()
		if host != "" && host != "localhost" && host != "127.0.0.1" && host != "postgres" {
			t.Skip("DATABASE_URL is not local; set a local DATABASE_URL for integration tests")
		}
	}
	pool, err := pgxpool.New(context.Background(), dsn)
	if err != nil {
		t.Fatalf("failed to connect to db: %v", err)
	}
	return pool
}

func withTx(t *testing.T, pool *pgxpool.Pool) (context.Context, repository.DBTX, *repository.Queries, func()) {
	t.Helper()
	ctx := context.Background()
	tx, err := pool.Begin(ctx)
	if err != nil {
		t.Fatalf("failed to begin tx: %v", err)
	}
	cleanup := func() {
		_ = tx.Rollback(ctx)
	}
	_, _ = tx.Exec(ctx, "select set_config('app.is_internal', 'true', false)")
	queries := repository.New(tx)
	return ctx, tx, queries, cleanup
}

func createTestUser(t *testing.T, ctx context.Context, queries *repository.Queries) session.UserInfo {
	t.Helper()
	userID := uuid.New()
	_, err := queries.CreateUserWithId(ctx, repository.CreateUserWithIdParams{
		ID:            userID,
		Name:          "Test User",
		Email:         "test@example.com",
		EmailVerified: true,
		Image:         nil,
	})
	if err != nil {
		t.Fatalf("failed to create user: %v", err)
	}
	return session.UserInfo{ID: userID.String(), Email: "test@example.com", Name: "Test User"}
}

func withUserContext(ctx context.Context, user session.UserInfo, queries *repository.Queries) context.Context {
	ctx = context.WithValue(ctx, session.UserContextKey, &user)
	ctx = WithQueries(ctx, queries)
	return ctx
}

func doJSONRequest(t *testing.T, handler http.HandlerFunc, method, path string, body any, ctx context.Context) *httptest.ResponseRecorder {
	t.Helper()
	var buf bytes.Buffer
	if body != nil {
		if err := json.NewEncoder(&buf).Encode(body); err != nil {
			t.Fatalf("encode body: %v", err)
		}
	}
	req := httptest.NewRequest(method, path, &buf).WithContext(ctx)
	req.Header.Set("Content-Type", "application/json")
	rec := httptest.NewRecorder()
	handler(rec, req)
	return rec
}

func makeLogger() *slog.Logger {
	return slog.New(slog.NewTextHandler(os.Stdout, nil))
}

func mustParseTime(t *testing.T, value string) time.Time {
	t.Helper()
	ts, err := time.Parse(time.RFC3339, value)
	if err != nil {
		t.Fatalf("parse time: %v", err)
	}
	return ts
}

func mustParseUUID(t *testing.T, value string) uuid.UUID {
	t.Helper()
	id, err := uuid.Parse(value)
	if err != nil {
		t.Fatalf("parse uuid: %v", err)
	}
	return id
}

func requireContains(t *testing.T, haystack, needle string) {
	t.Helper()
	if !strings.Contains(haystack, needle) {
		t.Fatalf("expected response to contain %q, got %q", needle, haystack)
	}
}
