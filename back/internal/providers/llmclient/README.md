# llmclient

```tree
llmclient/
├── README.md
└── llmclient.go
    ├── type Client {provider: llm.Provider}
    ├── func NewFromConfig(cfg config.Config) (*Client, error)
    ├── func NewWithProvider(provider llm.Provider) *Client
    ├── func (*Client) Name() string
    ├── func (*Client) Generate(ctx context.Context, systemPrompt string, messages []llm.Message) (string, error)
    └── func selectCredentials(provider string, cfg config.Config) (string, string, error)
```
