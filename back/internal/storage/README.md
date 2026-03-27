# storage

```tree
storage/
├── README.md
├── minio.go
│   ├── type MinIOStore {client: *minio.Client, bucketName: string, publicBase: string}
│   ├── func NewMinIOStore(cfg config.Config) (*MinIOStore, error)
│   ├── func (*MinIOStore) Put(ctx context.Context, key string, contentType string, data []byte) (string, error)
│   ├── func (*MinIOStore) Get(ctx context.Context, key string) ([]byte, error)
│   ├── func (*MinIOStore) ensureBucket(ctx context.Context) error
│   └── func (*MinIOStore) ensurePublicPolicy(ctx context.Context) error
└── storage.go
    └── type ObjectStore {Put: (ctx context.Context, key string, contentType string, data []byte) (string, error), Get: (ctx context.Context, key string) ([]byte, error)}
```
