# artifacts

```tree
artifacts/
├── README.md
└── store.go
    ├── type Store {root: string}
    ├── func New(root string) *Store
    ├── func (*Store) ensureRoot() error
    ├── func (*Store) Persist(sourceTag, artifactType, stage, identifier string, payload interface{}, metadata map[string]interface{}) (string, error)
    ├── func (*Store) Load(artifactID string) (map[string]interface{}, error)
    ├── func (*Store) List(limit int) ([]model.ReplayArtifact, error)
    └── func summaryFromPayload(payload map[string]interface{}) (model.ReplayArtifact, error)
```
