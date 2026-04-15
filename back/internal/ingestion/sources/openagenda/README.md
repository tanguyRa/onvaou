# openagenda

```tree
openagenda/
├── README.md
└── openagenda.go
    ├── type Service {cfg: config.Config, logger: *slog.Logger, client: *http.Client, geocoder: *geocoder.BANClient, store: *store.Store, artifacts: *artifacts.Store}
    ├── func NewService(cfg config.Config, logger *slog.Logger, client *http.Client, geocoder *geocoder.BANClient, store *store.Store, artifactStore *artifacts.Store) *Service
    ├── func (*Service) Run(ctx context.Context) (model.IngestionResult, error)
    ├── func (*Service) ReplayPayload(ctx context.Context, payload interface{}, metadata map[string]interface{}) (*model.Event, error)
    ├── func (*Service) syncAgendaCatalog(ctx context.Context) (map[string]int, error)
    ├── func (*Service) syncAgenda(ctx context.Context, agendaID int64) (model.IngestionResult, error)
    ├── func (*Service) fetchAgendaCatalog(ctx context.Context) ([]map[string]interface{}, error)
    ├── func (*Service) fetchAgendaEvents(ctx context.Context, agendaUID int64) ([]map[string]interface{}, error)
    ├── func (*Service) normalizeAgenda(raw map[string]interface{}) *model.OpenAgendaAgenda
    ├── func (*Service) normalizeEvents(ctx context.Context, rawEvents []map[string]interface{}, agendaUID int64) ([]model.Event, error)
    ├── func (*Service) normalizeOpenAgendaEvent(ctx context.Context, rawEvent map[string]interface{}, agendaUID int64) (*model.Event, error)
    ├── func (*Service) getJSON(ctx context.Context, path string, params url.Values, target interface{}) (int, error)
    ├── func extractCoordinates(payload map[string]interface{}) (float64, float64, bool)
    ├── func extractTiming(payload map[string]interface{}) (interface{}, interface{})
    ├── func boolToInt(value bool) int
    ├── func fallbackString(value string, fallback string) string
    ├── func cloneValues(values url.Values) url.Values
    ├── func firstNonNil(values ...interface{}) interface{}
    ├── func sumMapInts(values map[string]interface{}) int
    ├── func toInt(value interface{}) int
    ├── func toInt64(value interface{}) (int64, bool)
    ├── func toFloat64(value interface{}) (float64, bool)
    ├── func toBool(value interface{}) bool
    └── func anyString(value interface{}) string
```
