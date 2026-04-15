# model

```tree
model/
├── README.md
└── types.go
    ├── type Event {EventID: uuid.UUID, SourceUID: string, Title: string, Description: string, StartDT: time.Time, EndDT: *time.Time, LocationName: string, Address: string, Longitude: float64, Latitude: float64, SourceTag: string, SourceURL: string}
    ├── type IngestionResult {SourceTag: string, Fetched: int, Inserted: int, Updated: int, Duplicates: int, Failed: int, Details: []string}
    ├── type OpenAgendaAgenda {UID: int64, Slug: string, Title: string, Description: string, Official: bool, UpdatedAt: *time.Time, UpcomingEvents: int, RecentlyAddedEvents: int, SourceURL: string, LastPayloadHash: string, LastSeenAt: time.Time}
    ├── type ReplayArtifact {ArtifactID: string, SourceTag: string, ArtifactType: string, Stage: string, Identifier: string, SavedAt: time.Time, Metadata: map[string]interface{}}
    ├── func (Event) WithEventID() Event
    ├── func (Event) ContentHash() string
    ├── func (Event) DedupText() string
    ├── func NewResult(sourceTag string) IngestionResult
    └── func (*IngestionResult) Merge(other IngestionResult)
```
