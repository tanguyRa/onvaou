# datagouv

```tree
datagouv/
├── README.md
└── datagouv.go
    ├── type Service {cfg: config.Config, logger: *slog.Logger, client: *http.Client, geocoder: *geocoder.BANClient, store: *store.Store, artifacts: *artifacts.Store}
    ├── func NewService(cfg config.Config, logger *slog.Logger, client *http.Client, geocoder *geocoder.BANClient, store *store.Store, artifactStore *artifacts.Store) *Service
    ├── func (*Service) Run(ctx context.Context) (model.IngestionResult, error)
    ├── func (*Service) ReplayPayload(ctx context.Context, payload interface{}, metadata map[string]interface{}) (*model.Event, error)
    ├── func (*Service) fetchEvents(ctx context.Context) ([]model.Event, error)
    ├── func (*Service) parseResource(ctx context.Context, resource map[string]interface{}) ([]model.Event, error)
    ├── func (*Service) parseCSVResource(ctx context.Context, body io.Reader, sourceUIDPrefix string, sourceURL string) ([]model.Event, error)
    ├── func (*Service) parseJSONResource(ctx context.Context, body io.Reader, sourceUIDPrefix string, sourceURL string) ([]model.Event, error)
    ├── func (*Service) normalizeRow(ctx context.Context, row map[string]interface{}, sourceUID string, sourceURL string) (*model.Event, error)
    ├── func extractRows(payload interface{}) []map[string]interface{}
    ├── func matchesKeywords(values ...string) bool
    ├── func pickField(row map[string]interface{}, names ...string) string
    └── func fallbackString(value string, fallback string) string
```
