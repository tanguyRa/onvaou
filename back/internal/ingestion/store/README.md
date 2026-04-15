# store

```tree
store/
├── README.md
└── store.go
    ├── type Store {pool: *pgxpool.Pool, artifacts: *artifacts.Store}
    ├── func New(pool *pgxpool.Pool, artifactStore *artifacts.Store) *Store
    ├── func (*Store) UpsertEventBatch(ctx context.Context, sourceTag string, events []model.Event) (model.IngestionResult, error)
    ├── func (*Store) upsertEvent(ctx context.Context, tx pgx.Tx, event model.Event) (string, error)
    ├── func (*Store) UpsertOpenAgendaAgendas(ctx context.Context, agendas []model.OpenAgendaAgenda) (int, int, error)
    ├── func (*Store) StoreOpenAgendaAgendaFetch(ctx context.Context, batchHash string, rawAgendas []map[string]interface{}) (bool, error)
    ├── func (*Store) ListChangedOpenAgendaAgendas(ctx context.Context, agendas []model.OpenAgendaAgenda) ([]model.OpenAgendaAgenda, error)
    ├── func (*Store) StoreOpenAgendaAgendaPayloads(ctx context.Context, agendas []model.OpenAgendaAgenda) (int, error)
    ├── func (*Store) ListRelevantOpenAgendaAgendaIDs(ctx context.Context, limit int) ([]int64, error)
    ├── func (*Store) HasOpenAgendaEventBatchChanged(ctx context.Context, agendaUID int64, batchHash string) (bool, error)
    ├── func (*Store) StoreOpenAgendaEventFetch(ctx context.Context, agendaUID int64, batchHash string, rawEvents []map[string]interface{}) (bool, error)
    ├── func (*Store) MarkOpenAgendaAgendaSynced(ctx context.Context, agendaUID int64) error
    ├── func (*Store) MarkOpenAgendaAgendaSyncResult(ctx context.Context, agendaUID int64, batchHash string, fetchError string) error
    └── func (*Store) VacuumEvents(ctx context.Context) error
```
