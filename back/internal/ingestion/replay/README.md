# replay

```tree
replay/
├── README.md
└── replay.go
    ├── type openAgendaReplayer {ReplayPayload: (ctx context.Context, payload interface{}, metadata map[string]interface{}) (*model.Event, error)}
    ├── type dataGouvReplayer {ReplayPayload: (ctx context.Context, payload interface{}, metadata map[string]interface{}) (*model.Event, error)}
    ├── type Service {artifacts: *artifacts.Store, store: *store.Store, openAgenda: openAgendaReplayer, dataGouv: dataGouvReplayer}
    ├── func NewService(artifactStore *artifacts.Store, store *store.Store, openAgenda openAgendaReplayer, dataGouv dataGouvReplayer) *Service
    ├── func (*Service) ReplayArtifact(ctx context.Context, artifactID string) (model.IngestionResult, error)
    ├── func normalizedEventFromPayload(payload interface{}) (*model.Event, error)
    ├── func firstValue(raw map[string]interface{}, keys ...string) interface{}
    ├── func modelTime(value interface{}) *time.Time
    └── func modelFloat(value interface{}) float64
```
