# ingestion

```tree
ingestion/
├── README.md
├── runner.go
│   ├── type Runner {logger: *slog.Logger, openAgenda: *openagenda.Service, dataGouv: *datagouv.Service, replay: *replay.Service, store: *store.Store, artifactSet: *artifacts.Store}
│   ├── func NewRunner(cfg config.Config, logger *slog.Logger, pool *pgxpool.Pool) *Runner
│   ├── func (*Runner) RunSource(ctx context.Context, source string) (model.IngestionResult, error)
│   ├── func (*Runner) RunAll(ctx context.Context) ([]model.IngestionResult, error)
│   ├── func (*Runner) ReplayArtifact(ctx context.Context, artifactID string) (model.IngestionResult, error)
│   ├── func (*Runner) ListArtifacts(limit int) ([]model.ReplayArtifact, error)
│   └── func ValidateMode(source string, all bool, replayArtifact string, listArtifacts bool) error
├── artifacts/
│   ├── README.md
│   └── store.go
│       ├── type Store {root: string}
│       ├── func New(root string) *Store
│       ├── func (*Store) ensureRoot() error
│       ├── func (*Store) Persist(sourceTag, artifactType, stage, identifier string, payload interface{}, metadata map[string]interface{}) (string, error)
│       ├── func (*Store) Load(artifactID string) (map[string]interface{}, error)
│       ├── func (*Store) List(limit int) ([]model.ReplayArtifact, error)
│       └── func summaryFromPayload(payload map[string]interface{}) (model.ReplayArtifact, error)
├── dedup/
│   ├── README.md
│   └── dedup.go
│       ├── type Decision {EventID: uuid.UUID, Action: string}
│       ├── func CheckDuplicate(ctx context.Context, tx pgx.Tx, event model.Event) (Decision, error)
│       ├── func tokenSortRatio(left, right string) int
│       ├── func sortTokens(value string) string
│       ├── func levenshtein(left, right []rune) int
│       └── func min(values ...int) int
├── geocoder/
│   ├── README.md
│   └── ban.go
│       ├── type BANClient {baseURL: string, httpClient: *http.Client, retryAttempts: int}
│       ├── func NewBANClient(cfg config.Config, httpClient *http.Client) *BANClient
│       ├── func (*BANClient) ResolveAddress(ctx context.Context, address string) (float64, float64, error)
│       ├── func (*BANClient) ResolveCity(ctx context.Context, city string) (float64, float64, error)
│       ├── func (*BANClient) search(ctx context.Context, query string, limit int, banType string) (float64, float64, error)
│       └── func (*BANClient) doSearch(ctx context.Context, query string, limit int, banType string) (float64, float64, error)
├── model/
│   ├── README.md
│   └── types.go
│       ├── type Event {EventID: uuid.UUID, SourceUID: string, Title: string, Description: string, StartDT: time.Time, EndDT: *time.Time, LocationName: string, Address: string, Longitude: float64, Latitude: float64, SourceTag: string, SourceURL: string}
│       ├── type IngestionResult {SourceTag: string, Fetched: int, Inserted: int, Updated: int, Duplicates: int, Failed: int, Details: []string}
│       ├── type OpenAgendaAgenda {UID: int64, Slug: string, Title: string, Description: string, Official: bool, UpdatedAt: *time.Time, UpcomingEvents: int, RecentlyAddedEvents: int, SourceURL: string, LastPayloadHash: string, LastSeenAt: time.Time}
│       ├── type ReplayArtifact {ArtifactID: string, SourceTag: string, ArtifactType: string, Stage: string, Identifier: string, SavedAt: time.Time, Metadata: map[string]interface{}}
│       ├── func (Event) WithEventID() Event
│       ├── func (Event) ContentHash() string
│       ├── func (Event) DedupText() string
│       ├── func NewResult(sourceTag string) IngestionResult
│       └── func (*IngestionResult) Merge(other IngestionResult)
├── replay/
│   ├── README.md
│   └── replay.go
│       ├── type openAgendaReplayer {ReplayPayload: (ctx context.Context, payload interface{}, metadata map[string]interface{}) (*model.Event, error)}
│       ├── type dataGouvReplayer {ReplayPayload: (ctx context.Context, payload interface{}, metadata map[string]interface{}) (*model.Event, error)}
│       ├── type Service {artifacts: *artifacts.Store, store: *store.Store, openAgenda: openAgendaReplayer, dataGouv: dataGouvReplayer}
│       ├── func NewService(artifactStore *artifacts.Store, store *store.Store, openAgenda openAgendaReplayer, dataGouv dataGouvReplayer) *Service
│       ├── func (*Service) ReplayArtifact(ctx context.Context, artifactID string) (model.IngestionResult, error)
│       ├── func normalizedEventFromPayload(payload interface{}) (*model.Event, error)
│       ├── func firstValue(raw map[string]interface{}, keys ...string) interface{}
│       ├── func modelTime(value interface{}) *time.Time
│       └── func modelFloat(value interface{}) float64
├── sources/
│   ├── datagouv/
│   │   ├── README.md
│   │   └── datagouv.go
│   │       ├── type Service {cfg: config.Config, logger: *slog.Logger, client: *http.Client, geocoder: *geocoder.BANClient, store: *store.Store, artifacts: *artifacts.Store}
│   │       ├── func NewService(cfg config.Config, logger *slog.Logger, client *http.Client, geocoder *geocoder.BANClient, store *store.Store, artifactStore *artifacts.Store) *Service
│   │       ├── func (*Service) Run(ctx context.Context) (model.IngestionResult, error)
│   │       ├── func (*Service) ReplayPayload(ctx context.Context, payload interface{}, metadata map[string]interface{}) (*model.Event, error)
│   │       ├── func (*Service) fetchEvents(ctx context.Context) ([]model.Event, error)
│   │       ├── func (*Service) parseResource(ctx context.Context, resource map[string]interface{}) ([]model.Event, error)
│   │       ├── func (*Service) parseCSVResource(ctx context.Context, body io.Reader, sourceUIDPrefix string, sourceURL string) ([]model.Event, error)
│   │       ├── func (*Service) parseJSONResource(ctx context.Context, body io.Reader, sourceUIDPrefix string, sourceURL string) ([]model.Event, error)
│   │       ├── func (*Service) normalizeRow(ctx context.Context, row map[string]interface{}, sourceUID string, sourceURL string) (*model.Event, error)
│   │       ├── func extractRows(payload interface{}) []map[string]interface{}
│   │       ├── func matchesKeywords(values ...string) bool
│   │       ├── func pickField(row map[string]interface{}, names ...string) string
│   │       └── func fallbackString(value string, fallback string) string
│   └── openagenda/
│       ├── README.md
│       └── openagenda.go
│           ├── type Service {cfg: config.Config, logger: *slog.Logger, client: *http.Client, geocoder: *geocoder.BANClient, store: *store.Store, artifacts: *artifacts.Store}
│           ├── func NewService(cfg config.Config, logger *slog.Logger, client *http.Client, geocoder *geocoder.BANClient, store *store.Store, artifactStore *artifacts.Store) *Service
│           ├── func (*Service) Run(ctx context.Context) (model.IngestionResult, error)
│           ├── func (*Service) ReplayPayload(ctx context.Context, payload interface{}, metadata map[string]interface{}) (*model.Event, error)
│           ├── func (*Service) syncAgendaCatalog(ctx context.Context) (map[string]int, error)
│           ├── func (*Service) syncAgenda(ctx context.Context, agendaID int64) (model.IngestionResult, error)
│           ├── func (*Service) fetchAgendaCatalog(ctx context.Context) ([]map[string]interface{}, error)
│           ├── func (*Service) fetchAgendaEvents(ctx context.Context, agendaUID int64) ([]map[string]interface{}, error)
│           ├── func (*Service) normalizeAgenda(raw map[string]interface{}) *model.OpenAgendaAgenda
│           ├── func (*Service) normalizeEvents(ctx context.Context, rawEvents []map[string]interface{}, agendaUID int64) ([]model.Event, error)
│           ├── func (*Service) normalizeOpenAgendaEvent(ctx context.Context, rawEvent map[string]interface{}, agendaUID int64) (*model.Event, error)
│           ├── func (*Service) getJSON(ctx context.Context, path string, params url.Values, target interface{}) (int, error)
│           ├── func extractCoordinates(payload map[string]interface{}) (float64, float64, bool)
│           ├── func extractTiming(payload map[string]interface{}) (interface{}, interface{})
│           ├── func boolToInt(value bool) int
│           ├── func fallbackString(value string, fallback string) string
│           ├── func cloneValues(values url.Values) url.Values
│           ├── func firstNonNil(values ...interface{}) interface{}
│           ├── func sumMapInts(values map[string]interface{}) int
│           ├── func toInt(value interface{}) int
│           ├── func toInt64(value interface{}) (int64, bool)
│           ├── func toFloat64(value interface{}) (float64, bool)
│           ├── func toBool(value interface{}) bool
│           └── func anyString(value interface{}) string
├── store/
│   ├── README.md
│   └── store.go
│       ├── type Store {pool: *pgxpool.Pool, artifacts: *artifacts.Store}
│       ├── func New(pool *pgxpool.Pool, artifactStore *artifacts.Store) *Store
│       ├── func (*Store) UpsertEventBatch(ctx context.Context, sourceTag string, events []model.Event) (model.IngestionResult, error)
│       ├── func (*Store) upsertEvent(ctx context.Context, tx pgx.Tx, event model.Event) (string, error)
│       ├── func (*Store) UpsertOpenAgendaAgendas(ctx context.Context, agendas []model.OpenAgendaAgenda) (int, int, error)
│       ├── func (*Store) StoreOpenAgendaAgendaFetch(ctx context.Context, batchHash string, rawAgendas []map[string]interface{}) (bool, error)
│       ├── func (*Store) ListChangedOpenAgendaAgendas(ctx context.Context, agendas []model.OpenAgendaAgenda) ([]model.OpenAgendaAgenda, error)
│       ├── func (*Store) StoreOpenAgendaAgendaPayloads(ctx context.Context, agendas []model.OpenAgendaAgenda) (int, error)
│       ├── func (*Store) ListRelevantOpenAgendaAgendaIDs(ctx context.Context, limit int) ([]int64, error)
│       ├── func (*Store) HasOpenAgendaEventBatchChanged(ctx context.Context, agendaUID int64, batchHash string) (bool, error)
│       ├── func (*Store) StoreOpenAgendaEventFetch(ctx context.Context, agendaUID int64, batchHash string, rawEvents []map[string]interface{}) (bool, error)
│       ├── func (*Store) MarkOpenAgendaAgendaSynced(ctx context.Context, agendaUID int64) error
│       ├── func (*Store) MarkOpenAgendaAgendaSyncResult(ctx context.Context, agendaUID int64, batchHash string, fetchError string) error
│       └── func (*Store) VacuumEvents(ctx context.Context) error
└── util/
    ├── README.md
    └── util.go
        ├── func NormalizeSpace(value string) string
        ├── func FirstString(value interface{}) string
        ├── func BuildAddress(parts ...string) string
        ├── func ParseDateTime(value interface{}) *time.Time
        └── func PayloadHash(value interface{}) string
```
