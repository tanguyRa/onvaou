# config

```tree
config/
├── README.md
└── config.go
    ├── type Config {Environment: string, Address: string, Encryption: EncryptionConfig, Database: DatabaseConfig, Ingestion: IngestionConfig, Payment: PaymentConfig, Polar: PolarConfig, LLM: LLMsConfig, Storage: StorageConfig}
    ├── type IngestionConfig {BANAPIURL: string, DataGouvAPIURL: string, OpenAgendaAPIURL: string, OpenAgendaAPIKey: string, OpenAgendaOfficialOnly: bool, OpenAgendaAgendaUpdatedWithinDays: int, OpenAgendaMaxAgendas: int, HTTPTimeoutSeconds: int, BANRetryAttempts: int, DebugDir: string, SchedulerTimezone: string, OpenAgendaQueries: []string, DataGouvQueries: []string}
    ├── type LLMsConfig {Provider: string, Google: LLMConfig, OpenAI: LLMConfig, Anthropic: LLMConfig}
    ├── type LLMConfig {APIKey: string, Model: string}
    ├── type StorageConfig {Provider: string, MinIO: MinIOConfig}
    ├── type MinIOConfig {Endpoint: string, AccessKey: string, SecretKey: string, Bucket: string, UseSSL: bool, PublicBase: string}
    ├── type PaymentConfig {Provider: string, Stripe: StripeConfig, Polar: PolarConfig}
    ├── type StripeConfig {APIKey: string}
    ├── type PolarConfig {WebhookSecret: string}
    ├── type EncryptionConfig {Key: string}
    ├── type DatabaseConfig {ConnectionString: string}
    ├── func Load() (*Config, error)
    ├── func loadFromFile(path string, config *Config) error
    ├── func loadFromEnv(config *Config)
    ├── func parseInt(s string) (int, error)
    ├── func parseBool(s string) (bool, error)
    ├── func setDefaults() *Config
    ├── func validate(config *Config) error
    └── func splitCSV(value string) []string
```
