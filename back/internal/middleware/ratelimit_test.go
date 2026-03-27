package middleware

import (
	"net/http"
	"net/http/httptest"
	"testing"
	"time"
)

func TestRateLimiter_Allow(t *testing.T) {
	// Create a rate limiter: 2 requests per second, burst of 3
	rl := NewRateLimiter(2, time.Second, 3)

	key := "test-user"

	// First 3 requests should be allowed (burst)
	for i := 0; i < 3; i++ {
		if !rl.Allow(key) {
			t.Errorf("request %d should be allowed (within burst)", i+1)
		}
	}

	// 4th request should be denied
	if rl.Allow(key) {
		t.Error("4th request should be denied (burst exceeded)")
	}
}

func TestRateLimiter_Refill(t *testing.T) {
	// Create a rate limiter: 10 requests per 100ms, burst of 2
	rl := NewRateLimiter(10, 100*time.Millisecond, 2)

	key := "test-user"

	// Exhaust the burst
	for i := 0; i < 2; i++ {
		rl.Allow(key)
	}

	// Should be denied
	if rl.Allow(key) {
		t.Error("should be denied after burst exhausted")
	}

	// Wait for refill
	time.Sleep(110 * time.Millisecond)

	// Should be allowed after refill
	if !rl.Allow(key) {
		t.Error("should be allowed after refill")
	}
}

func TestRateLimiter_DifferentKeys(t *testing.T) {
	rl := NewRateLimiter(1, time.Minute, 2)

	// User 1 exhausts their limit
	rl.Allow("user1")
	rl.Allow("user1")

	// User 1 should be denied
	if rl.Allow("user1") {
		t.Error("user1 should be denied")
	}

	// User 2 should still be allowed
	if !rl.Allow("user2") {
		t.Error("user2 should be allowed (separate limit)")
	}
}

func TestRateLimitMiddleware(t *testing.T) {
	rl := NewRateLimiter(1, time.Minute, 2)

	keyExtractor := func(r *http.Request) string {
		return r.Header.Get("X-User-ID")
	}

	handler := http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusOK)
		w.Write([]byte("OK"))
	})

	middleware := RateLimitMiddleware(rl, keyExtractor)(handler)

	// First 2 requests should succeed
	for i := 0; i < 2; i++ {
		req := httptest.NewRequest("GET", "/test", nil)
		req.Header.Set("X-User-ID", "user123")
		rec := httptest.NewRecorder()

		middleware.ServeHTTP(rec, req)

		if rec.Code != http.StatusOK {
			t.Errorf("request %d: expected 200, got %d", i+1, rec.Code)
		}
	}

	// 3rd request should be rate limited
	req := httptest.NewRequest("GET", "/test", nil)
	req.Header.Set("X-User-ID", "user123")
	rec := httptest.NewRecorder()

	middleware.ServeHTTP(rec, req)

	if rec.Code != http.StatusTooManyRequests {
		t.Errorf("expected 429, got %d", rec.Code)
	}
}

func TestRateLimitMiddleware_FallbackToIP(t *testing.T) {
	rl := NewRateLimiter(1, time.Minute, 1)

	// Key extractor returns empty string
	keyExtractor := func(r *http.Request) string {
		return ""
	}

	handler := http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusOK)
	})

	middleware := RateLimitMiddleware(rl, keyExtractor)(handler)

	// First request from same IP succeeds
	req := httptest.NewRequest("GET", "/test", nil)
	req.RemoteAddr = "192.168.1.1:12345"
	rec := httptest.NewRecorder()

	middleware.ServeHTTP(rec, req)

	if rec.Code != http.StatusOK {
		t.Errorf("expected 200, got %d", rec.Code)
	}

	// Second request from same IP is rate limited
	req2 := httptest.NewRequest("GET", "/test", nil)
	req2.RemoteAddr = "192.168.1.1:12345"
	rec2 := httptest.NewRecorder()

	middleware.ServeHTTP(rec2, req2)

	if rec2.Code != http.StatusTooManyRequests {
		t.Errorf("expected 429, got %d", rec2.Code)
	}

	// Request from different IP succeeds
	req3 := httptest.NewRequest("GET", "/test", nil)
	req3.RemoteAddr = "192.168.1.2:12345"
	rec3 := httptest.NewRecorder()

	middleware.ServeHTTP(rec3, req3)

	if rec3.Code != http.StatusOK {
		t.Errorf("expected 200, got %d", rec3.Code)
	}
}

func TestRateLimitMiddleware_ResponseHeaders(t *testing.T) {
	rl := NewRateLimiter(1, time.Minute, 1)
	defer rl.Stop()

	keyExtractor := func(r *http.Request) string {
		return "user123"
	}

	handler := http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusOK)
	})

	middleware := RateLimitMiddleware(rl, keyExtractor)(handler)

	// Exhaust rate limit
	req1 := httptest.NewRequest("GET", "/test", nil)
	rec1 := httptest.NewRecorder()
	middleware.ServeHTTP(rec1, req1)

	// Trigger rate limit
	req2 := httptest.NewRequest("GET", "/test", nil)
	rec2 := httptest.NewRecorder()
	middleware.ServeHTTP(rec2, req2)

	if rec2.Code != http.StatusTooManyRequests {
		t.Errorf("expected 429, got %d", rec2.Code)
	}

	// Check Content-Type is JSON
	if ct := rec2.Header().Get("Content-Type"); ct != "application/json" {
		t.Errorf("expected Content-Type application/json, got %q", ct)
	}

	// Check Retry-After header
	if ra := rec2.Header().Get("Retry-After"); ra != "60" {
		t.Errorf("expected Retry-After 60, got %q", ra)
	}
}

func TestGetClientIP(t *testing.T) {
	tests := []struct {
		name       string
		xff        string
		xri        string
		remoteAddr string
		want       string
	}{
		{
			name:       "X-Forwarded-For single IP",
			xff:        "203.0.113.195",
			remoteAddr: "10.0.0.1:12345",
			want:       "203.0.113.195",
		},
		{
			name:       "X-Forwarded-For multiple IPs",
			xff:        "203.0.113.195, 70.41.3.18, 150.172.238.178",
			remoteAddr: "10.0.0.1:12345",
			want:       "203.0.113.195",
		},
		{
			name:       "X-Real-IP",
			xri:        "203.0.113.195",
			remoteAddr: "10.0.0.1:12345",
			want:       "203.0.113.195",
		},
		{
			name:       "Fallback to RemoteAddr with port",
			remoteAddr: "192.168.1.100:54321",
			want:       "192.168.1.100",
		},
		{
			name:       "Fallback to RemoteAddr without port",
			remoteAddr: "192.168.1.100",
			want:       "192.168.1.100",
		},
		{
			name:       "X-Forwarded-For takes precedence over X-Real-IP",
			xff:        "203.0.113.195",
			xri:        "70.41.3.18",
			remoteAddr: "10.0.0.1:12345",
			want:       "203.0.113.195",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			req := httptest.NewRequest("GET", "/test", nil)
			req.RemoteAddr = tt.remoteAddr
			if tt.xff != "" {
				req.Header.Set("X-Forwarded-For", tt.xff)
			}
			if tt.xri != "" {
				req.Header.Set("X-Real-IP", tt.xri)
			}

			got := getClientIP(req)
			if got != tt.want {
				t.Errorf("getClientIP() = %q, want %q", got, tt.want)
			}
		})
	}
}

func TestRateLimiter_Stop(t *testing.T) {
	rl := NewRateLimiter(1, time.Second, 1)

	// Verify it works before stop
	if !rl.Allow("test") {
		t.Error("should allow before stop")
	}

	// Stop the rate limiter
	rl.Stop()

	// Give the goroutine time to exit
	time.Sleep(10 * time.Millisecond)

	// Rate limiter should still function after stop
	// (Stop only stops the cleanup goroutine, not the rate limiting)
	rl.Allow("test2") // Should not panic
}

func BenchmarkRateLimiter_Allow(b *testing.B) {
	rl := NewRateLimiter(1000, time.Second, 100)
	defer rl.Stop()

	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		rl.Allow("user-" + string(rune(i%100)))
	}
}
