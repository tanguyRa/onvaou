# geocoder

```tree
geocoder/
├── README.md
└── ban.go
    ├── type BANClient {baseURL: string, httpClient: *http.Client, retryAttempts: int}
    ├── func NewBANClient(cfg config.Config, httpClient *http.Client) *BANClient
    ├── func (*BANClient) ResolveAddress(ctx context.Context, address string) (float64, float64, error)
    ├── func (*BANClient) ResolveCity(ctx context.Context, city string) (float64, float64, error)
    ├── func (*BANClient) search(ctx context.Context, query string, limit int, banType string) (float64, float64, error)
    └── func (*BANClient) doSearch(ctx context.Context, query string, limit int, banType string) (float64, float64, error)
```
