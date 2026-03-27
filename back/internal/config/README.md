# config

```tree
config/
├── README.md
└── config.go
    ├── type Config {Environment: string, Address: string, Encryption: EncryptionConfig, Database: DatabaseConfig, Payment: PaymentConfig, Polar: PolarConfig, LLM: LLMsConfig, Storage: StorageConfig}
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
    └── func validate(config *Config) error
```
