package datagouv

import (
	"context"
	"encoding/csv"
	"encoding/json"
	"fmt"
	"io"
	"log/slog"
	"net/http"
	"net/url"
	"strconv"
	"strings"

	"github.com/tanguyRa/onvaou/internal/config"
	"github.com/tanguyRa/onvaou/internal/ingestion/artifacts"
	"github.com/tanguyRa/onvaou/internal/ingestion/geocoder"
	"github.com/tanguyRa/onvaou/internal/ingestion/model"
	"github.com/tanguyRa/onvaou/internal/ingestion/store"
	"github.com/tanguyRa/onvaou/internal/ingestion/util"
)

var (
	datasetKeywords = []string{"agenda", "evenement", "événement", "manifestation", "culture", "festival", "mairie"}
)

type Service struct {
	cfg       config.Config
	logger    *slog.Logger
	client    *http.Client
	geocoder  *geocoder.BANClient
	store     *store.Store
	artifacts *artifacts.Store
}

func NewService(cfg config.Config, logger *slog.Logger, client *http.Client, geocoder *geocoder.BANClient, store *store.Store, artifactStore *artifacts.Store) *Service {
	return &Service{
		cfg:       cfg,
		logger:    logger,
		client:    client,
		geocoder:  geocoder,
		store:     store,
		artifacts: artifactStore,
	}
}

func (s *Service) Run(ctx context.Context) (model.IngestionResult, error) {
	events, err := s.fetchEvents(ctx)
	if err != nil {
		return model.IngestionResult{}, err
	}
	return s.store.UpsertEventBatch(ctx, "datagouv", events)
}

func (s *Service) ReplayPayload(ctx context.Context, payload interface{}, metadata map[string]interface{}) (*model.Event, error) {
	row, ok := payload.(map[string]interface{})
	if !ok {
		return nil, fmt.Errorf("data.gouv replay payload must be an object")
	}
	sourceUID := fmt.Sprint(metadata["source_uid"])
	sourceURL := fmt.Sprint(metadata["source_url"])
	return s.normalizeRow(ctx, row, sourceUID, sourceURL)
}

func (s *Service) fetchEvents(ctx context.Context) ([]model.Event, error) {
	endpoint, err := url.Parse(s.cfg.Ingestion.DataGouvAPIURL)
	if err != nil {
		return nil, err
	}

	queries := s.cfg.Ingestion.DataGouvQueries
	if len(queries) == 0 {
		queries = []string{"agenda", "evenement", "culture", "festival", "mairie"}
	}

	events := make([]model.Event, 0)
	seenDatasetIDs := map[string]struct{}{}
	seenResourceIDs := map[string]struct{}{}
	totalDatasets := 0

	for _, query := range queries {
		searchURL := *endpoint
		searchURL.Path += "/datasets/"
		params := searchURL.Query()
		params.Set("page_size", "20")
		params.Set("q", query)
		searchURL.RawQuery = params.Encode()

		req, err := http.NewRequestWithContext(ctx, http.MethodGet, searchURL.String(), nil)
		if err != nil {
			return nil, err
		}

		resp, err := s.client.Do(req)
		if err != nil {
			return nil, err
		}
		if resp.StatusCode >= 400 {
			resp.Body.Close()
			return nil, fmt.Errorf("data.gouv upstream returned status %d", resp.StatusCode)
		}

		var payload map[string]interface{}
		err = json.NewDecoder(resp.Body).Decode(&payload)
		resp.Body.Close()
		if err != nil {
			return nil, err
		}

		rawDatasets := extractRows(payload)
		totalDatasets += len(rawDatasets)

		for _, dataset := range rawDatasets {
			datasetID := fallbackString(fmt.Sprint(dataset["id"]), fmt.Sprint(dataset["slug"]))
			if datasetID != "" {
				if _, exists := seenDatasetIDs[datasetID]; exists {
					continue
				}
				seenDatasetIDs[datasetID] = struct{}{}
			}

			title := util.FirstString(dataset["title"])
			description := util.FirstString(dataset["description"])
			if !matchesKeywords(title, description) {
				continue
			}

			resources, _ := dataset["resources"].([]interface{})
			for _, rawResource := range resources {
				resource, ok := rawResource.(map[string]interface{})
				if !ok {
					continue
				}
				resourceID := fallbackString(fmt.Sprint(resource["id"]), fmt.Sprint(resource["url"]))
				if resourceID != "" {
					if _, exists := seenResourceIDs[resourceID]; exists {
						continue
					}
					seenResourceIDs[resourceID] = struct{}{}
				}

				formatName := strings.ToLower(util.FirstString(resource["format"]))
				if formatName != "csv" && formatName != "json" && formatName != "geojson" {
					continue
				}

				parsedEvents, err := s.parseResource(ctx, resource)
				if err != nil {
					s.logger.Warn("skipping data.gouv resource", "resource", resourceID, "error", err)
					continue
				}
				events = append(events, parsedEvents...)
			}
		}
	}

	if totalDatasets == 0 {
		s.logger.Warn("data.gouv search returned no datasets", "queries", strings.Join(queries, ","))
	}

	return events, nil
}

func (s *Service) parseResource(ctx context.Context, resource map[string]interface{}) ([]model.Event, error) {
	rawURL, ok := resource["url"].(string)
	if !ok || rawURL == "" {
		return nil, nil
	}

	req, err := http.NewRequestWithContext(ctx, http.MethodGet, rawURL, nil)
	if err != nil {
		return nil, err
	}

	resp, err := s.client.Do(req)
	if err != nil {
		return nil, err
	}
	defer resp.Body.Close()

	if resp.StatusCode >= 400 {
		return nil, fmt.Errorf("resource upstream returned status %d", resp.StatusCode)
	}

	formatName := strings.ToLower(util.FirstString(resource["format"]))
	sourceUIDPrefix := fallbackString(fmt.Sprint(resource["id"]), rawURL)

	switch formatName {
	case "csv":
		return s.parseCSVResource(ctx, resp.Body, sourceUIDPrefix, rawURL)
	case "json", "geojson":
		return s.parseJSONResource(ctx, resp.Body, sourceUIDPrefix, rawURL)
	default:
		contentType := resp.Header.Get("Content-Type")
		if strings.Contains(strings.ToLower(contentType), "json") {
			return s.parseJSONResource(ctx, resp.Body, sourceUIDPrefix, rawURL)
		}
	}

	return nil, nil
}

func (s *Service) parseCSVResource(ctx context.Context, body io.Reader, sourceUIDPrefix string, sourceURL string) ([]model.Event, error) {
	reader := csv.NewReader(body)
	rows, err := reader.ReadAll()
	if err != nil {
		return nil, err
	}
	if len(rows) == 0 {
		return nil, nil
	}

	headers := rows[0]
	events := make([]model.Event, 0, len(rows)-1)
	for index, rowValues := range rows[1:] {
		row := map[string]interface{}{}
		for i, header := range headers {
			if i < len(rowValues) {
				row[header] = rowValues[i]
			}
		}
		sourceUID := fmt.Sprintf("%s:%d", sourceUIDPrefix, index)
		event, err := s.normalizeRow(ctx, row, sourceUID, sourceURL)
		if err != nil {
			artifactID, artifactErr := s.artifacts.Persist(
				"datagouv",
				"datagouv_raw",
				"normalize",
				sourceUID,
				row,
				map[string]interface{}{
					"error":      err.Error(),
					"source_uid": sourceUID,
					"source_url": sourceURL,
				},
			)
			if artifactErr == nil {
				s.logger.Warn("skipping data.gouv record", "source_uid", sourceUID, "error", err, "artifact_id", artifactID)
			} else {
				s.logger.Warn("skipping data.gouv record", "source_uid", sourceUID, "error", err)
			}
			continue
		}
		if event != nil {
			events = append(events, *event)
		}
	}
	return events, nil
}

func (s *Service) parseJSONResource(ctx context.Context, body io.Reader, sourceUIDPrefix string, sourceURL string) ([]model.Event, error) {
	var payload interface{}
	if err := json.NewDecoder(body).Decode(&payload); err != nil {
		return nil, err
	}

	rows := extractRows(payload)
	events := make([]model.Event, 0, len(rows))
	for index, row := range rows {
		sourceUID := fmt.Sprintf("%s:%d", sourceUIDPrefix, index)
		event, err := s.normalizeRow(ctx, row, sourceUID, sourceURL)
		if err != nil {
			artifactID, artifactErr := s.artifacts.Persist(
				"datagouv",
				"datagouv_raw",
				"normalize",
				sourceUID,
				row,
				map[string]interface{}{
					"error":      err.Error(),
					"source_uid": sourceUID,
					"source_url": sourceURL,
				},
			)
			if artifactErr == nil {
				s.logger.Warn("skipping data.gouv record", "source_uid", sourceUID, "error", err, "artifact_id", artifactID)
			} else {
				s.logger.Warn("skipping data.gouv record", "source_uid", sourceUID, "error", err)
			}
			continue
		}
		if event != nil {
			events = append(events, *event)
		}
	}
	return events, nil
}

func (s *Service) normalizeRow(ctx context.Context, row map[string]interface{}, sourceUID string, sourceURL string) (*model.Event, error) {
	title := pickField(row, "title", "titre", "name", "nom", "event", "summary")
	startDT := util.ParseDateTime(
		pickField(row, "start_dt", "start_date", "date_start", "date_debut", "date", "datetime"),
	)
	address := util.BuildAddress(
		pickField(row, "address", "adresse", "lieu", "location"),
		pickField(row, "postal_code", "code_postal", "postcode"),
		pickField(row, "city", "ville", "commune"),
	)
	if title == "" || startDT == nil || address == "" {
		return nil, nil
	}

	lonText := pickField(row, "longitude", "lon", "lng", "x")
	latText := pickField(row, "latitude", "lat", "y")

	var lon float64
	var lat float64
	if lonText != "" && latText != "" {
		var err error
		lon, err = strconv.ParseFloat(strings.ReplaceAll(lonText, ",", "."), 64)
		if err != nil {
			return nil, err
		}
		lat, err = strconv.ParseFloat(strings.ReplaceAll(latText, ",", "."), 64)
		if err != nil {
			return nil, err
		}
	} else {
		var err error
		lon, lat, err = s.geocoder.ResolveAddress(ctx, address)
		if err != nil {
			return nil, err
		}
	}

	return &model.Event{
		SourceUID:    util.NormalizeSpace(sourceUID),
		Title:        title,
		Description:  pickField(row, "description", "details", "resume", "summary"),
		StartDT:      *startDT,
		EndDT:        util.ParseDateTime(pickField(row, "end_dt", "end_date", "date_fin", "end")),
		LocationName: pickField(row, "location_name", "venue", "place", "lieu"),
		Address:      address,
		Longitude:    lon,
		Latitude:     lat,
		SourceTag:    "datagouv",
		SourceURL:    sourceURL,
	}, nil
}

func extractRows(payload interface{}) []map[string]interface{} {
	switch typed := payload.(type) {
	case []interface{}:
		out := make([]map[string]interface{}, 0, len(typed))
		for _, item := range typed {
			if row, ok := item.(map[string]interface{}); ok {
				out = append(out, row)
			}
		}
		return out
	case map[string]interface{}:
		for _, key := range []string{"data", "results", "records", "items"} {
			if rows, ok := typed[key].([]interface{}); ok {
				out := make([]map[string]interface{}, 0, len(rows))
				for _, item := range rows {
					if row, ok := item.(map[string]interface{}); ok {
						out = append(out, row)
					}
				}
				return out
			}
		}
		return []map[string]interface{}{typed}
	default:
		return nil
	}
}

func matchesKeywords(values ...string) bool {
	haystack := strings.ToLower(strings.Join(values, " "))
	for _, keyword := range datasetKeywords {
		if strings.Contains(haystack, keyword) {
			return true
		}
	}
	return false
}

func pickField(row map[string]interface{}, names ...string) string {
	lowered := make(map[string]string, len(row))
	for key := range row {
		lowered[strings.ToLower(key)] = key
	}
	for _, name := range names {
		if key, ok := lowered[name]; ok {
			return util.FirstString(row[key])
		}
	}
	return ""
}

func fallbackString(value string, fallback string) string {
	if value == "" {
		return fallback
	}
	return value
}
