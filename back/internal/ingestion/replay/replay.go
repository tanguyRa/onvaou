package replay

import (
	"context"
	"fmt"
	"time"

	"github.com/tanguyRa/onvaou/internal/ingestion/artifacts"
	"github.com/tanguyRa/onvaou/internal/ingestion/model"
	"github.com/tanguyRa/onvaou/internal/ingestion/store"
	"github.com/tanguyRa/onvaou/internal/ingestion/util"
)

type openAgendaReplayer interface {
	ReplayPayload(ctx context.Context, payload interface{}, metadata map[string]interface{}) (*model.Event, error)
}

type dataGouvReplayer interface {
	ReplayPayload(ctx context.Context, payload interface{}, metadata map[string]interface{}) (*model.Event, error)
}

type Service struct {
	artifacts  *artifacts.Store
	store      *store.Store
	openAgenda openAgendaReplayer
	dataGouv   dataGouvReplayer
}

func NewService(artifactStore *artifacts.Store, store *store.Store, openAgenda openAgendaReplayer, dataGouv dataGouvReplayer) *Service {
	return &Service{
		artifacts:  artifactStore,
		store:      store,
		openAgenda: openAgenda,
		dataGouv:   dataGouv,
	}
}

func (s *Service) ReplayArtifact(ctx context.Context, artifactID string) (model.IngestionResult, error) {
	artifact, err := s.artifacts.Load(artifactID)
	if err != nil {
		return model.IngestionResult{}, err
	}

	artifactType := fmt.Sprint(artifact["artifact_type"])
	payload := artifact["payload"]
	metadata, _ := artifact["metadata"].(map[string]interface{})

	var event *model.Event
	switch artifactType {
	case "openagenda_raw":
		event, err = s.openAgenda.ReplayPayload(ctx, payload, metadata)
	case "datagouv_raw":
		event, err = s.dataGouv.ReplayPayload(ctx, payload, metadata)
	case "normalized_event":
		event, err = normalizedEventFromPayload(payload)
	default:
		return model.IngestionResult{}, fmt.Errorf("unsupported artifact type %q", artifactType)
	}
	if err != nil {
		return model.IngestionResult{}, err
	}
	if event == nil {
		result := model.NewResult("replay")
		result.Details = append(result.Details, "replay payload still does not normalize into an event")
		return result, nil
	}

	return s.store.UpsertEventBatch(ctx, event.SourceTag, []model.Event{*event})
}

func normalizedEventFromPayload(payload interface{}) (*model.Event, error) {
	raw, ok := payload.(map[string]interface{})
	if !ok {
		return nil, fmt.Errorf("normalized_event payload must be an object")
	}

	startDT := raw["StartDT"]
	if startDT == nil {
		startDT = raw["start_dt"]
	}
	parsedStart := modelTime(startDT)
	if parsedStart == nil {
		return nil, fmt.Errorf("normalized_event missing start_dt")
	}

	return &model.Event{
		SourceUID:    fmt.Sprint(firstValue(raw, "SourceUID", "source_uid")),
		Title:        fmt.Sprint(firstValue(raw, "Title", "title")),
		Description:  fmt.Sprint(firstValue(raw, "Description", "description")),
		StartDT:      *parsedStart,
		EndDT:        modelTime(firstValue(raw, "EndDT", "end_dt")),
		LocationName: fmt.Sprint(firstValue(raw, "LocationName", "location_name")),
		Address:      fmt.Sprint(firstValue(raw, "Address", "address")),
		Longitude:    modelFloat(firstValue(raw, "Longitude", "longitude")),
		Latitude:     modelFloat(firstValue(raw, "Latitude", "latitude")),
		SourceTag:    fmt.Sprint(firstValue(raw, "SourceTag", "source_tag")),
		SourceURL:    fmt.Sprint(firstValue(raw, "SourceURL", "source_url")),
	}, nil
}

func firstValue(raw map[string]interface{}, keys ...string) interface{} {
	for _, key := range keys {
		if value, ok := raw[key]; ok {
			return value
		}
	}
	return nil
}

func modelTime(value interface{}) *time.Time {
	return util.ParseDateTime(value)
}

func modelFloat(value interface{}) float64 {
	switch typed := value.(type) {
	case float64:
		return typed
	case float32:
		return float64(typed)
	case int:
		return float64(typed)
	case int64:
		return float64(typed)
	case string:
		var out float64
		fmt.Sscanf(typed, "%f", &out)
		return out
	default:
		return 0
	}
}
