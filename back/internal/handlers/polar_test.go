package handlers

import (
	"bytes"
	"crypto/hmac"
	"crypto/sha256"
	"encoding/base64"
	"encoding/json"
	"fmt"
	"net/http"
	"net/http/httptest"
	"testing"
	"time"

	"github.com/tanguyRa/onvaou/internal/config"
)

func TestPolarWebhook(t *testing.T) {
	pool := requireLocalDB(t)
	defer pool.Close()

	ctx, _, queries, cleanup := withTx(t, pool)
	defer cleanup()

	handler := NewPolarHandler(queries, makeLogger(), config.PolarConfig{WebhookSecret: "whsec_" + base64.StdEncoding.EncodeToString([]byte("secret"))})

	event := WebhookEvent{
		Type:      "unknown.event",
		Timestamp: time.Now().UTC(),
		Data:      json.RawMessage(`{"hello":"world"}`),
	}
	body, err := json.Marshal(event)
	if err != nil {
		t.Fatalf("marshal event: %v", err)
	}

	webhookID := "wh_123"
	timestamp := fmt.Sprintf("%d", time.Now().Unix())
	signedContent := fmt.Sprintf("%s.%s.%s", webhookID, timestamp, string(body))
	mac := hmac.New(sha256.New, []byte("secret"))
	mac.Write([]byte(signedContent))
	signature := base64.StdEncoding.EncodeToString(mac.Sum(nil))

	req := httptest.NewRequest(http.MethodPost, "/webhooks/polar", bytes.NewReader(body)).WithContext(ctx)
	req.Header.Set(HeaderWebhookID, webhookID)
	req.Header.Set(HeaderWebhookTimestamp, timestamp)
	req.Header.Set(HeaderWebhookSignature, "v1,"+signature)
	rec := httptest.NewRecorder()

	handler.HandleWebhook(rec, req)
	if rec.Code != http.StatusOK {
		t.Fatalf("expected 200, got %d: %s", rec.Code, rec.Body.String())
	}
}
