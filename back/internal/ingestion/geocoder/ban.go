package geocoder

import (
	"context"
	"encoding/json"
	"fmt"
	"net/http"
	"net/url"
	"time"

	"github.com/tanguyRa/onvaou/internal/config"
)

type BANClient struct {
	baseURL       string
	httpClient    *http.Client
	retryAttempts int
}

func NewBANClient(cfg config.Config, httpClient *http.Client) *BANClient {
	return &BANClient{
		baseURL:       cfg.Ingestion.BANAPIURL,
		httpClient:    httpClient,
		retryAttempts: cfg.Ingestion.BANRetryAttempts,
	}
}

func (c *BANClient) ResolveAddress(ctx context.Context, address string) (float64, float64, error) {
	return c.search(ctx, address, 1, "")
}

func (c *BANClient) ResolveCity(ctx context.Context, city string) (float64, float64, error) {
	return c.search(ctx, city, 1, "municipality")
}

func (c *BANClient) search(ctx context.Context, query string, limit int, banType string) (float64, float64, error) {
	if query == "" {
		return 0, 0, fmt.Errorf("query is required")
	}

	var lastErr error
	attempts := c.retryAttempts
	if attempts < 1 {
		attempts = 1
	}

	for attempt := 1; attempt <= attempts; attempt++ {
		lon, lat, err := c.doSearch(ctx, query, limit, banType)
		if err == nil {
			return lon, lat, nil
		}
		lastErr = err
		if attempt < attempts {
			select {
			case <-ctx.Done():
				return 0, 0, ctx.Err()
			case <-time.After(time.Duration(1<<(attempt-1)) * time.Second):
			}
		}
	}

	return 0, 0, fmt.Errorf("failed to search BAN for query %q: %w", query, lastErr)
}

func (c *BANClient) doSearch(ctx context.Context, query string, limit int, banType string) (float64, float64, error) {
	endpoint, err := url.Parse(c.baseURL)
	if err != nil {
		return 0, 0, err
	}
	endpoint.Path += "/search/"
	params := endpoint.Query()
	params.Set("q", query)
	params.Set("limit", fmt.Sprintf("%d", limit))
	params.Set("autocomplete", "1")
	if banType != "" {
		params.Set("type", banType)
	}
	endpoint.RawQuery = params.Encode()

	req, err := http.NewRequestWithContext(ctx, http.MethodGet, endpoint.String(), nil)
	if err != nil {
		return 0, 0, err
	}

	resp, err := c.httpClient.Do(req)
	if err != nil {
		return 0, 0, err
	}
	defer resp.Body.Close()

	if resp.StatusCode >= 400 {
		return 0, 0, fmt.Errorf("ban upstream returned status %d", resp.StatusCode)
	}

	var payload struct {
		Features []struct {
			Geometry struct {
				Coordinates []float64 `json:"coordinates"`
			} `json:"geometry"`
		} `json:"features"`
	}
	if err := json.NewDecoder(resp.Body).Decode(&payload); err != nil {
		return 0, 0, err
	}
	if len(payload.Features) == 0 || len(payload.Features[0].Geometry.Coordinates) < 2 {
		return 0, 0, fmt.Errorf("ban returned no coordinates")
	}

	return payload.Features[0].Geometry.Coordinates[0], payload.Features[0].Geometry.Coordinates[1], nil
}
