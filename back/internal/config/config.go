package config

import (
	"encoding/base64"
	"encoding/json"
	"fmt"
	"os"
)

type Config struct {
	Environment string           `json:"environment"`
	Address     string           `json:"address"`
	Encryption  EncryptionConfig `json:"encryption"`
	Database    DatabaseConfig   `json:"database"`
	Payment     PaymentConfig    `json:"payment"`
	Polar       PolarConfig      `json:"polar"`
	LLM         LLMsConfig       `json:"llm"`
	Storage     StorageConfig    `json:"storage"`
}

type LLMsConfig struct {
	Provider string `json:"provider"`

	Google    LLMConfig `json:"google"`
	OpenAI    LLMConfig `json:"openai"`
	Anthropic LLMConfig `json:"anthropic"`
}
type LLMConfig struct {
	APIKey string `json:"apiKey"`
	Model  string `json:"model"`
}

type StorageConfig struct {
	Provider string `json:"provider"`

	MinIO MinIOConfig `json:"minio"`
}
type MinIOConfig struct {
	Endpoint   string `json:"minioEndpoint"`
	AccessKey  string `json:"minioAccessKey"`
	SecretKey  string `json:"minioSecretKey"`
	Bucket     string `json:"minioBucket"`
	UseSSL     bool   `json:"minioUseSsl"`
	PublicBase string `json:"minioPublicBase"`
}

type PaymentConfig struct {
	Provider string `json:"provider"`

	Stripe StripeConfig `json:"stripe"`
	Polar  PolarConfig  `json:"polar"`
}
type StripeConfig struct {
	APIKey string
}
type PolarConfig struct {
	WebhookSecret string `json:"webhookSecret"`
}

type EncryptionConfig struct {
	Key string `json:"key"`
}

type DatabaseConfig struct {
	ConnectionString string `json:"connectionString"`
}

// Load reads configuration from environment variables and optionally from a config file
func Load() (*Config, error) {
	// Start with defaults
	config := setDefaults()

	// Override with file config if present
	if configPath := os.Getenv("CONFIG_FILE"); configPath != "" {
		if err := loadFromFile(configPath, config); err != nil {
			return nil, fmt.Errorf("error loading config file: %w", err)
		}
	}

	// Override with environment variables
	loadFromEnv(config)

	// Validate final configuration
	if err := validate(config); err != nil {
		return nil, fmt.Errorf("invalid configuration: %w", err)
	}

	return config, nil
}

func loadFromFile(path string, config *Config) error {
	file, err := os.Open(path)
	if err != nil {
		return err
	}
	defer file.Close()

	return json.NewDecoder(file).Decode(config)
}

func loadFromEnv(config *Config) {
	if environment := os.Getenv("ENVIRONMENT"); environment != "" {
		config.Environment = environment
	}

	if address := os.Getenv("ADDRESS"); address != "" {
		config.Address = address
	}

	// Encryption configuration
	if encryptionKey := os.Getenv("ENCRYPTION_KEY"); encryptionKey != "" {
		config.Encryption.Key = encryptionKey
	}

	// Database configuration
	if databaseURL := os.Getenv("DATABASE_URL"); databaseURL != "" {
		config.Database = DatabaseConfig{
			ConnectionString: databaseURL,
		}
	}

	// Polar configuration
	if polarWebhookSecret := os.Getenv("POLAR_WEBHOOK_SECRET"); polarWebhookSecret != "" {
		config.Polar.WebhookSecret = polarWebhookSecret
	}

	// LLMs configuration
	if LLM_PROVIDER := os.Getenv("LLM_PROVIDER"); LLM_PROVIDER != "" {
		config.LLM.Provider = LLM_PROVIDER
	}

	// LLM configuration
	if llmProvider := os.Getenv("LLM_PROVIDER"); llmProvider != "" {
		config.LLM.Provider = llmProvider
	}
	if geminiAPIKey := os.Getenv("GEMINI_API_KEY"); geminiAPIKey != "" {
		config.LLM.Google.APIKey = geminiAPIKey
	}
	if geminiModel := os.Getenv("GEMINI_MODEL"); geminiModel != "" {
		config.LLM.Google.Model = geminiModel
	}
	if openAIAPIKey := os.Getenv("OPENAI_API_KEY"); openAIAPIKey != "" {
		config.LLM.OpenAI.APIKey = openAIAPIKey
	}
	if openAIModel := os.Getenv("OPENAI_MODEL"); openAIModel != "" {
		config.LLM.OpenAI.Model = openAIModel
	}
	if anthropicAPIKey := os.Getenv("ANTHROPIC_API_KEY"); anthropicAPIKey != "" {
		config.LLM.Anthropic.APIKey = anthropicAPIKey
	}
	if anthropicModel := os.Getenv("ANTHROPIC_MODEL"); anthropicModel != "" {
		config.LLM.Anthropic.Model = anthropicModel
	}

	// Storage configuration
	if storageProvider := os.Getenv("STORAGE_PROVIDER"); storageProvider != "" {
		config.Storage.Provider = storageProvider
	}
	if minioEndpoint := os.Getenv("MINIO_ENDPOINT"); minioEndpoint != "" {
		config.Storage.MinIO.Endpoint = minioEndpoint
	}
	if minioAccessKey := os.Getenv("MINIO_ACCESS_KEY"); minioAccessKey != "" {
		config.Storage.MinIO.AccessKey = minioAccessKey
	}
	if minioSecretKey := os.Getenv("MINIO_SECRET_KEY"); minioSecretKey != "" {
		config.Storage.MinIO.SecretKey = minioSecretKey
	}
	if minioBucket := os.Getenv("MINIO_BUCKET"); minioBucket != "" {
		config.Storage.MinIO.Bucket = minioBucket
	}
	if minioUseSSL := os.Getenv("MINIO_USE_SSL"); minioUseSSL != "" {
		if v, err := parseBool(minioUseSSL); err == nil {
			config.Storage.MinIO.UseSSL = v
		}
	}
	if minioPublicBase := os.Getenv("MINIO_PUBLIC_BASE_URL"); minioPublicBase != "" {
		config.Storage.MinIO.PublicBase = minioPublicBase
	}
}

func parseInt(s string) (int, error) {
	var v int
	_, err := fmt.Sscanf(s, "%d", &v)
	return v, err
}

func parseBool(s string) (bool, error) {
	var v bool
	_, err := fmt.Sscanf(s, "%t", &v)
	return v, err
}

func setDefaults() *Config {
	return &Config{
		Environment: "production",
		Address:     ":8080",
		LLM: LLMsConfig{
			Provider: "google",
			Google:   LLMConfig{Model: "gemini-3.0-flash"},
		},
		Storage: StorageConfig{
			Provider: "fs",
		},
	}
}

func validate(config *Config) error {
	if config.Environment == "" {
		return fmt.Errorf("environment is required")
	}

	if config.Address == "" {
		return fmt.Errorf("address is required")
	}

	// Encryption key validation
	if config.Encryption.Key != "" {
		// Decode base64 key and check decoded length
		decodedKey, err := base64.StdEncoding.DecodeString(config.Encryption.Key)
		if err != nil {
			return fmt.Errorf("encryption key must be valid base64: %w", err)
		}
		if len(decodedKey) != 32 {
			return fmt.Errorf("encryption key must decode to exactly 32 bytes (256 bits), got %d bytes", len(decodedKey))
		}
	}

	// Database configuration validation
	if config.Database.ConnectionString == "" {
		return fmt.Errorf("database connection string is required")
	}

	return nil
}
