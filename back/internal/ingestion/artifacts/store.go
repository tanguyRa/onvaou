package artifacts

import (
	"encoding/json"
	"fmt"
	"os"
	"path/filepath"
	"sort"
	"time"

	"github.com/google/uuid"
	"github.com/tanguyRa/onvaou/internal/ingestion/model"
)

type Store struct {
	root string
}

func New(root string) *Store {
	return &Store{root: root}
}

func (s *Store) ensureRoot() error {
	return os.MkdirAll(s.root, 0o755)
}

func (s *Store) Persist(sourceTag, artifactType, stage, identifier string, payload interface{}, metadata map[string]interface{}) (string, error) {
	if err := s.ensureRoot(); err != nil {
		return "", err
	}

	savedAt := time.Now().UTC()
	artifactID := fmt.Sprintf("%s-%s", savedAt.Format("20060102T150405"), uuid.NewString())
	envelope := map[string]interface{}{
		"artifact_id":   artifactID,
		"source_tag":    sourceTag,
		"artifact_type": artifactType,
		"stage":         stage,
		"identifier":    identifier,
		"saved_at":      savedAt.Format(time.RFC3339Nano),
		"metadata":      metadata,
		"payload":       payload,
	}

	data, err := json.MarshalIndent(envelope, "", "  ")
	if err != nil {
		return "", err
	}

	path := filepath.Join(s.root, artifactID+".json")
	if err := os.WriteFile(path, data, 0o644); err != nil {
		return "", err
	}

	return artifactID, nil
}

func (s *Store) Load(artifactID string) (map[string]interface{}, error) {
	path := filepath.Join(s.root, artifactID+".json")
	data, err := os.ReadFile(path)
	if err != nil {
		return nil, err
	}

	var payload map[string]interface{}
	if err := json.Unmarshal(data, &payload); err != nil {
		return nil, err
	}

	return payload, nil
}

func (s *Store) List(limit int) ([]model.ReplayArtifact, error) {
	if err := s.ensureRoot(); err != nil {
		return nil, err
	}

	entries, err := os.ReadDir(s.root)
	if err != nil {
		return nil, err
	}
	sort.Slice(entries, func(i, j int) bool {
		return entries[i].Name() > entries[j].Name()
	})

	if limit <= 0 {
		limit = 50
	}

	out := make([]model.ReplayArtifact, 0, limit)
	for _, entry := range entries {
		if entry.IsDir() {
			continue
		}
		if len(out) >= limit {
			break
		}
		artifactID := entry.Name()
		if filepath.Ext(artifactID) != ".json" {
			continue
		}
		artifactID = artifactID[:len(artifactID)-len(".json")]
		payload, err := s.Load(artifactID)
		if err != nil {
			return nil, err
		}
		artifact, err := summaryFromPayload(payload)
		if err != nil {
			return nil, err
		}
		out = append(out, artifact)
	}

	return out, nil
}

func summaryFromPayload(payload map[string]interface{}) (model.ReplayArtifact, error) {
	savedAt, err := time.Parse(time.RFC3339Nano, fmt.Sprint(payload["saved_at"]))
	if err != nil {
		return model.ReplayArtifact{}, err
	}

	metadata := map[string]interface{}{}
	if rawMetadata, ok := payload["metadata"].(map[string]interface{}); ok {
		metadata = rawMetadata
	}

	return model.ReplayArtifact{
		ArtifactID:   fmt.Sprint(payload["artifact_id"]),
		SourceTag:    fmt.Sprint(payload["source_tag"]),
		ArtifactType: fmt.Sprint(payload["artifact_type"]),
		Stage:        fmt.Sprint(payload["stage"]),
		Identifier:   fmt.Sprint(payload["identifier"]),
		SavedAt:      savedAt,
		Metadata:     metadata,
	}, nil
}
