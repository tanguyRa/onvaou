package openagenda

import (
	"context"
	"encoding/json"
	"fmt"
	"log/slog"
	"net/http"
	"net/url"
	"strconv"
	"time"

	"github.com/tanguyRa/onvaou/internal/config"
	"github.com/tanguyRa/onvaou/internal/ingestion/artifacts"
	"github.com/tanguyRa/onvaou/internal/ingestion/geocoder"
	"github.com/tanguyRa/onvaou/internal/ingestion/model"
	"github.com/tanguyRa/onvaou/internal/ingestion/store"
	"github.com/tanguyRa/onvaou/internal/ingestion/util"
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
	catalogResult, err := s.syncAgendaCatalog(ctx)
	if err != nil {
		return model.IngestionResult{}, err
	}

	agendaIDs, err := s.store.ListRelevantOpenAgendaAgendaIDs(ctx, s.cfg.Ingestion.OpenAgendaMaxAgendas)
	if err != nil {
		return model.IngestionResult{}, err
	}

	result := model.NewResult("openagenda")
	result.Details = append(result.Details, fmt.Sprintf(
		"agenda catalog: fetched=%d changed=%d inserted=%d updated=%d",
		catalogResult["fetched"],
		catalogResult["changed"],
		catalogResult["inserted"],
		catalogResult["updated"],
	))

	for _, agendaID := range agendaIDs {
		agendaResult, err := s.syncAgenda(ctx, agendaID)
		if err != nil {
			result.Failed++
			result.Details = append(result.Details, fmt.Sprintf("agenda %d: %v", agendaID, err))
			continue
		}
		result.Merge(agendaResult)
	}

	return result, nil
}

func (s *Service) ReplayPayload(ctx context.Context, payload interface{}, metadata map[string]interface{}) (*model.Event, error) {
	raw, ok := payload.(map[string]interface{})
	if !ok {
		return nil, fmt.Errorf("openagenda replay payload must be an object")
	}

	var agendaUID int64
	if metadata != nil {
		switch typed := metadata["agenda_uid"].(type) {
		case float64:
			agendaUID = int64(typed)
		case int64:
			agendaUID = typed
		case string:
			if parsed, err := strconv.ParseInt(typed, 10, 64); err == nil {
				agendaUID = parsed
			}
		}
	}

	return s.normalizeOpenAgendaEvent(ctx, raw, agendaUID)
}

func (s *Service) syncAgendaCatalog(ctx context.Context) (map[string]int, error) {
	rawAgendas, err := s.fetchAgendaCatalog(ctx)
	if err != nil {
		return nil, err
	}

	batchHash := util.PayloadHash(rawAgendas)
	isNewFetch, err := s.store.StoreOpenAgendaAgendaFetch(ctx, batchHash, rawAgendas)
	if err != nil {
		return nil, err
	}

	agendas := make([]model.OpenAgendaAgenda, 0, len(rawAgendas))
	for _, rawAgenda := range rawAgendas {
		agenda := s.normalizeAgenda(rawAgenda)
		if agenda != nil {
			agendas = append(agendas, *agenda)
		}
	}

	changedAgendas, err := s.store.ListChangedOpenAgendaAgendas(ctx, agendas)
	if err != nil {
		return nil, err
	}

	storedPayloads, err := s.store.StoreOpenAgendaAgendaPayloads(ctx, changedAgendas)
	if err != nil {
		return nil, err
	}

	inserted, updated, err := s.store.UpsertOpenAgendaAgendas(ctx, changedAgendas)
	if err != nil {
		return nil, err
	}

	return map[string]int{
		"fetched":         len(rawAgendas),
		"batch_new":       boolToInt(isNewFetch),
		"changed":         len(changedAgendas),
		"payloads_stored": storedPayloads,
		"inserted":        inserted,
		"updated":         updated,
	}, nil
}

func (s *Service) syncAgenda(ctx context.Context, agendaID int64) (model.IngestionResult, error) {
	rawEvents, err := s.fetchAgendaEvents(ctx, agendaID)
	if err != nil {
		_ = s.store.MarkOpenAgendaAgendaSyncResult(ctx, agendaID, "", err.Error())
		return model.IngestionResult{}, err
	}

	batchHash := util.PayloadHash(rawEvents)
	if _, err := s.store.StoreOpenAgendaEventFetch(ctx, agendaID, batchHash, rawEvents); err != nil {
		_ = s.store.MarkOpenAgendaAgendaSyncResult(ctx, agendaID, "", err.Error())
		return model.IngestionResult{}, err
	}

	changed, err := s.store.HasOpenAgendaEventBatchChanged(ctx, agendaID, batchHash)
	if err != nil {
		_ = s.store.MarkOpenAgendaAgendaSyncResult(ctx, agendaID, "", err.Error())
		return model.IngestionResult{}, err
	}
	if !changed {
		result := model.NewResult("openagenda")
		result.Fetched = len(rawEvents)
		result.Details = append(result.Details, fmt.Sprintf("agenda %d: unchanged batch", agendaID))
		_ = s.store.MarkOpenAgendaAgendaSyncResult(ctx, agendaID, batchHash, "")
		_ = s.store.MarkOpenAgendaAgendaSynced(ctx, agendaID)
		return result, nil
	}

	events, err := s.normalizeEvents(ctx, rawEvents, agendaID)
	if err != nil {
		_ = s.store.MarkOpenAgendaAgendaSyncResult(ctx, agendaID, "", err.Error())
		return model.IngestionResult{}, err
	}

	result, err := s.store.UpsertEventBatch(ctx, "openagenda", events)
	if err != nil {
		_ = s.store.MarkOpenAgendaAgendaSyncResult(ctx, agendaID, "", err.Error())
		return model.IngestionResult{}, err
	}
	if err := s.store.MarkOpenAgendaAgendaSyncResult(ctx, agendaID, batchHash, ""); err != nil {
		return model.IngestionResult{}, err
	}
	if err := s.store.MarkOpenAgendaAgendaSynced(ctx, agendaID); err != nil {
		return model.IngestionResult{}, err
	}

	result.Details = append(result.Details, fmt.Sprintf("agenda %d: processed", agendaID))
	return result, nil
}

func (s *Service) fetchAgendaCatalog(ctx context.Context) ([]map[string]interface{}, error) {
	if s.cfg.Ingestion.OpenAgendaAPIKey == "" {
		s.logger.Warn("OPENAGENDA_API_KEY is not configured; skipping agenda discovery")
		return nil, nil
	}

	params := url.Values{}
	params.Set("size", "100")
	params.Set("sort", "recentlyAddedEvents.desc")
	for _, field := range []string{"description", "slug", "summary", "title", "uid", "updatedAt", "official"} {
		params.Add("includeFields[]", field)
	}
	if s.cfg.Ingestion.OpenAgendaOfficialOnly {
		params.Set("official", "1")
	}
	if days := s.cfg.Ingestion.OpenAgendaAgendaUpdatedWithinDays; days > 0 {
		params.Set("updatedAt.gte", time.Now().UTC().AddDate(0, 0, -days).Format(time.RFC3339))
	}

	var agendas []map[string]interface{}
	var after []interface{}
	page := 0

	for {
		page++
		requestParams := cloneValues(params)
		for _, item := range after {
			requestParams.Add("after[]", fmt.Sprint(item))
		}

		var payload struct {
			Agendas []map[string]interface{} `json:"agendas"`
			After   []interface{}            `json:"after"`
		}
		statusCode, err := s.getJSON(ctx, "/agendas", requestParams, &payload)
		if err != nil {
			if statusCode >= 500 && len(agendas) > 0 {
				s.logger.Warn("openagenda agenda pagination stopped after upstream failure", "page", page, "status", statusCode, "collected", len(agendas))
				break
			}
			return nil, err
		}

		agendas = append(agendas, payload.Agendas...)
		after = payload.After
		if len(after) == 0 {
			break
		}
	}

	return agendas, nil
}

func (s *Service) fetchAgendaEvents(ctx context.Context, agendaUID int64) ([]map[string]interface{}, error) {
	if s.cfg.Ingestion.OpenAgendaAPIKey == "" {
		s.logger.Warn("OPENAGENDA_API_KEY is not configured; skipping agenda event fetch")
		return nil, nil
	}

	var events []map[string]interface{}
	var after []interface{}

	for {
		params := url.Values{}
		params.Set("size", "300")
		params.Set("monolingual", "fr")
		params.Set("detailed", "1")
		params.Add("relative[]", "current")
		params.Add("relative[]", "upcoming")
		for _, item := range after {
			params.Add("after[]", fmt.Sprint(item))
		}

		var payload struct {
			Events []map[string]interface{} `json:"events"`
			After  []interface{}            `json:"after"`
		}
		_, err := s.getJSON(ctx, fmt.Sprintf("/agendas/%d/events", agendaUID), params, &payload)
		if err != nil {
			return nil, err
		}

		events = append(events, payload.Events...)
		after = payload.After
		if len(after) == 0 {
			break
		}
	}

	return events, nil
}

func (s *Service) normalizeAgenda(raw map[string]interface{}) *model.OpenAgendaAgenda {
	uidValue, ok := raw["uid"]
	if !ok {
		return nil
	}
	uid, ok := toInt64(uidValue)
	if !ok {
		return nil
	}

	summary, _ := raw["summary"].(map[string]interface{})
	publishedEvents, _ := summary["publishedEvents"].(map[string]interface{})
	recentlyAdded, _ := summary["recentlyAddedEvents"].(map[string]interface{})

	var updatedAt *time.Time
	if parsed := util.ParseDateTime(raw["updatedAt"]); parsed != nil {
		updatedAt = parsed
	}

	return &model.OpenAgendaAgenda{
		UID:                 uid,
		Slug:                util.FirstString(raw["slug"]),
		Title:               fallbackString(util.FirstString(raw["title"]), strconv.FormatInt(uid, 10)),
		Description:         util.FirstString(raw["description"]),
		Official:            toBool(raw["official"]),
		UpdatedAt:           updatedAt,
		UpcomingEvents:      toInt(publishedEvents["current"]) + toInt(publishedEvents["upcoming"]),
		RecentlyAddedEvents: sumMapInts(recentlyAdded),
		SourceURL:           fmt.Sprintf("https://openagenda.com/agendas/%d", uid),
		LastPayloadHash:     util.PayloadHash(raw),
		LastSeenAt:          time.Now().UTC(),
	}
}

func (s *Service) normalizeEvents(ctx context.Context, rawEvents []map[string]interface{}, agendaUID int64) ([]model.Event, error) {
	events := make([]model.Event, 0, len(rawEvents))
	for _, rawEvent := range rawEvents {
		event, err := s.normalizeOpenAgendaEvent(ctx, rawEvent, agendaUID)
		if err != nil {
			identifier := fmt.Sprintf("%d:%s", agendaUID, fallbackString(anyString(rawEvent["uid"]), fallbackString(anyString(rawEvent["id"]), "unknown")))
			artifactID, artifactErr := s.artifacts.Persist(
				"openagenda",
				"openagenda_raw",
				"normalize",
				identifier,
				rawEvent,
				map[string]interface{}{
					"error":      err.Error(),
					"agenda_uid": agendaUID,
				},
			)
			if artifactErr == nil {
				s.logger.Warn("skipping openagenda event", "agenda_uid", agendaUID, "identifier", identifier, "error", err, "artifact_id", artifactID)
			} else {
				s.logger.Warn("skipping openagenda event", "agenda_uid", agendaUID, "identifier", identifier, "error", err)
			}
			continue
		}
		if event != nil {
			events = append(events, *event)
		}
	}
	return events, nil
}

func (s *Service) normalizeOpenAgendaEvent(ctx context.Context, rawEvent map[string]interface{}, agendaUID int64) (*model.Event, error) {
	title := util.FirstString(rawEvent["title"])
	startRaw, endRaw := extractTiming(rawEvent)
	startDT := util.ParseDateTime(startRaw)

	location, _ := rawEvent["location"].(map[string]interface{})
	address := util.BuildAddress(
		util.FirstString(rawEvent["address"]),
		util.FirstString(location["address"]),
		util.FirstString(location["postalCode"]),
		util.FirstString(location["city"]),
	)
	if title == "" || startDT == nil || address == "" {
		return nil, nil
	}

	lon, lat, ok := extractCoordinates(rawEvent)
	if !ok {
		var err error
		lon, lat, err = s.geocoder.ResolveAddress(ctx, address)
		if err != nil {
			return nil, err
		}
	}

	rawUID := fallbackString(anyString(rawEvent["uid"]), fallbackString(anyString(rawEvent["id"]), fallbackString(util.FirstString(rawEvent["slug"]), title)))
	sourceUID := rawUID
	if agendaUID != 0 {
		sourceUID = fmt.Sprintf("%d:%s", agendaUID, rawUID)
	}

	sourceURL := util.FirstString(rawEvent["canonicalUrl"])
	if sourceURL == "" {
		sourceURL = util.FirstString(rawEvent["url"])
	}
	if sourceURL == "" {
		sourceURL = util.FirstString(rawEvent["html"])
	}
	if sourceURL == "" {
		if agendaUID != 0 {
			sourceURL = fmt.Sprintf("https://openagenda.com/agendas/%d/events/%s", agendaUID, rawUID)
		} else {
			sourceURL = fmt.Sprintf("https://openagenda.com/%s", rawUID)
		}
	}

	return &model.Event{
		SourceUID:    util.NormalizeSpace(sourceUID),
		Title:        title,
		Description:  fallbackString(util.FirstString(rawEvent["longDescription"]), util.FirstString(rawEvent["description"])),
		StartDT:      *startDT,
		EndDT:        util.ParseDateTime(endRaw),
		LocationName: fallbackString(util.FirstString(location["name"]), util.FirstString(rawEvent["locationName"])),
		Address:      address,
		Longitude:    lon,
		Latitude:     lat,
		SourceTag:    "openagenda",
		SourceURL:    sourceURL,
	}, nil
}

func (s *Service) getJSON(ctx context.Context, path string, params url.Values, target interface{}) (int, error) {
	endpoint, err := url.Parse(s.cfg.Ingestion.OpenAgendaAPIURL)
	if err != nil {
		return 0, err
	}
	endpoint.Path += path
	endpoint.RawQuery = params.Encode()

	req, err := http.NewRequestWithContext(ctx, http.MethodGet, endpoint.String(), nil)
	if err != nil {
		return 0, err
	}
	req.Header.Set("key", s.cfg.Ingestion.OpenAgendaAPIKey)

	resp, err := s.client.Do(req)
	if err != nil {
		return 0, err
	}
	defer resp.Body.Close()

	if resp.StatusCode >= 400 {
		return resp.StatusCode, fmt.Errorf("openagenda upstream returned status %d", resp.StatusCode)
	}

	return resp.StatusCode, json.NewDecoder(resp.Body).Decode(target)
}

func extractCoordinates(payload map[string]interface{}) (float64, float64, bool) {
	if geom, ok := payload["geom"].(map[string]interface{}); ok {
		if coords, ok := geom["coordinates"].([]interface{}); ok && len(coords) >= 2 {
			lon, lonOK := toFloat64(coords[0])
			lat, latOK := toFloat64(coords[1])
			if lonOK && latOK {
				return lon, lat, true
			}
		}
	}
	if location, ok := payload["location"].(map[string]interface{}); ok {
		if coords, ok := location["coordinates"].([]interface{}); ok && len(coords) >= 2 {
			lon, lonOK := toFloat64(coords[0])
			lat, latOK := toFloat64(coords[1])
			if lonOK && latOK {
				return lon, lat, true
			}
		}
	}
	lon, lonOK := toFloat64(payload["longitude"])
	if !lonOK {
		lon, lonOK = toFloat64(payload["lon"])
	}
	lat, latOK := toFloat64(payload["latitude"])
	if !latOK {
		lat, latOK = toFloat64(payload["lat"])
	}
	if lonOK && latOK {
		return lon, lat, true
	}
	return 0, 0, false
}

func extractTiming(payload map[string]interface{}) (interface{}, interface{}) {
	for _, key := range []string{"timings", "timing", "dates"} {
		rawValue, ok := payload[key]
		if !ok {
			continue
		}
		items, ok := rawValue.([]interface{})
		if !ok || len(items) == 0 {
			continue
		}
		firstItem, ok := items[0].(map[string]interface{})
		if !ok {
			continue
		}
		start := firstNonNil(firstItem["begin"], firstItem["start"], firstItem["startTime"])
		end := firstNonNil(firstItem["end"], firstItem["finish"], firstItem["endTime"])
		return start, end
	}
	return payload["firstDate"], payload["lastDate"]
}

func boolToInt(value bool) int {
	if value {
		return 1
	}
	return 0
}

func fallbackString(value string, fallback string) string {
	if value == "" {
		return fallback
	}
	return value
}

func cloneValues(values url.Values) url.Values {
	out := url.Values{}
	for key, items := range values {
		for _, item := range items {
			out.Add(key, item)
		}
	}
	return out
}

func firstNonNil(values ...interface{}) interface{} {
	for _, value := range values {
		if value != nil {
			return value
		}
	}
	return nil
}

func sumMapInts(values map[string]interface{}) int {
	total := 0
	for _, value := range values {
		total += toInt(value)
	}
	return total
}

func toInt(value interface{}) int {
	switch typed := value.(type) {
	case float64:
		return int(typed)
	case int:
		return typed
	case int64:
		return int(typed)
	case json.Number:
		v, _ := typed.Int64()
		return int(v)
	case string:
		parsed, _ := strconv.Atoi(typed)
		return parsed
	default:
		return 0
	}
}

func toInt64(value interface{}) (int64, bool) {
	switch typed := value.(type) {
	case float64:
		return int64(typed), true
	case int64:
		return typed, true
	case int:
		return int64(typed), true
	case json.Number:
		v, err := typed.Int64()
		return v, err == nil
	case string:
		v, err := strconv.ParseInt(typed, 10, 64)
		return v, err == nil
	default:
		return 0, false
	}
}

func toFloat64(value interface{}) (float64, bool) {
	switch typed := value.(type) {
	case float64:
		return typed, true
	case float32:
		return float64(typed), true
	case int:
		return float64(typed), true
	case int64:
		return float64(typed), true
	case json.Number:
		v, err := typed.Float64()
		return v, err == nil
	case string:
		v, err := strconv.ParseFloat(typed, 64)
		return v, err == nil
	default:
		return 0, false
	}
}

func toBool(value interface{}) bool {
	switch typed := value.(type) {
	case bool:
		return typed
	case string:
		return typed == "1" || typed == "true" || typed == "True"
	default:
		return false
	}
}

func anyString(value interface{}) string {
	if text := util.FirstString(value); text != "" {
		return text
	}
	if value == nil {
		return ""
	}
	return fmt.Sprint(value)
}
