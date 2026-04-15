package model

import (
	"crypto/sha256"
	"encoding/hex"
	"strings"
	"time"

	"github.com/google/uuid"
)

type Event struct {
	EventID      uuid.UUID
	SourceUID    string
	Title        string
	Description  string
	StartDT      time.Time
	EndDT        *time.Time
	LocationName string
	Address      string
	Longitude    float64
	Latitude     float64
	SourceTag    string
	SourceURL    string
}

func (e Event) WithEventID() Event {
	if e.EventID != uuid.Nil {
		return e
	}

	basis := e.SourceTag + ":" + e.SourceUID + ":" + e.StartDT.UTC().Format(time.RFC3339Nano)
	e.EventID = uuid.NewSHA1(uuid.NameSpaceURL, []byte(basis))
	return e
}

func (e Event) ContentHash() string {
	payload := strings.ToLower(strings.TrimSpace(e.Title)) + "|" + e.StartDT.UTC().Format(time.RFC3339Nano) + "|" + strings.ToLower(strings.TrimSpace(e.SourceTag))
	sum := sha256.Sum256([]byte(payload))
	return hex.EncodeToString(sum[:])
}

func (e Event) DedupText() string {
	parts := []string{
		strings.ToLower(strings.TrimSpace(e.Title)),
		strings.ToLower(strings.TrimSpace(e.Address)),
	}
	return strings.TrimSpace(strings.Join(parts, " "))
}

type IngestionResult struct {
	SourceTag  string   `json:"source_tag"`
	Fetched    int      `json:"fetched"`
	Inserted   int      `json:"inserted"`
	Updated    int      `json:"updated"`
	Duplicates int      `json:"duplicates"`
	Failed     int      `json:"failed"`
	Details    []string `json:"details"`
}

func NewResult(sourceTag string) IngestionResult {
	return IngestionResult{
		SourceTag: sourceTag,
		Details:   make([]string, 0),
	}
}

func (r *IngestionResult) Merge(other IngestionResult) {
	r.Fetched += other.Fetched
	r.Inserted += other.Inserted
	r.Updated += other.Updated
	r.Duplicates += other.Duplicates
	r.Failed += other.Failed
	r.Details = append(r.Details, other.Details...)
}

type OpenAgendaAgenda struct {
	UID                 int64
	Slug                string
	Title               string
	Description         string
	Official            bool
	UpdatedAt           *time.Time
	UpcomingEvents      int
	RecentlyAddedEvents int
	SourceURL           string
	LastPayloadHash     string
	LastSeenAt          time.Time
}

type ReplayArtifact struct {
	ArtifactID   string                 `json:"artifact_id"`
	SourceTag    string                 `json:"source_tag"`
	ArtifactType string                 `json:"artifact_type"`
	Stage        string                 `json:"stage"`
	Identifier   string                 `json:"identifier"`
	SavedAt      time.Time              `json:"saved_at"`
	Metadata     map[string]interface{} `json:"metadata"`
}
