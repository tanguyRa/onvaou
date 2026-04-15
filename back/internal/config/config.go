package config

import (
	"encoding/base64"
	"encoding/json"
	"fmt"
	"os"
	"strings"
)

type Config struct {
	Environment string           `json:"environment"`
	Address     string           `json:"address"`
	Encryption  EncryptionConfig `json:"encryption"`
	Database    DatabaseConfig   `json:"database"`
	Ingestion   IngestionConfig  `json:"ingestion"`
	Payment     PaymentConfig    `json:"payment"`
	Polar       PolarConfig      `json:"polar"`
	LLM         LLMsConfig       `json:"llm"`
	Storage     StorageConfig    `json:"storage"`
}

type IngestionConfig struct {
	BANAPIURL                         string   `json:"banApiUrl"`
	DataGouvAPIURL                    string   `json:"datagouvApiUrl"`
	OpenAgendaAPIURL                  string   `json:"openagendaApiUrl"`
	OpenAgendaAPIKey                  string   `json:"openagendaApiKey"`
	OpenAgendaOfficialOnly            bool     `json:"openagendaOfficialOnly"`
	OpenAgendaAgendaUpdatedWithinDays int      `json:"openagendaAgendaUpdatedWithinDays"`
	OpenAgendaMaxAgendas              int      `json:"openagendaMaxAgendas"`
	HTTPTimeoutSeconds                int      `json:"httpTimeoutSeconds"`
	BANRetryAttempts                  int      `json:"banRetryAttempts"`
	DebugDir                          string   `json:"debugDir"`
	SchedulerTimezone                 string   `json:"schedulerTimezone"`
	OpenAgendaQueries                 []string `json:"openagendaQueries"`
	DataGouvQueries                   []string `json:"datagouvQueries"`
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

	// Ingestion configuration
	if value := os.Getenv("BAN_API_URL"); value != "" {
		config.Ingestion.BANAPIURL = value
	}
	if value := os.Getenv("DATAGOUV_API_URL"); value != "" {
		config.Ingestion.DataGouvAPIURL = value
	}
	if value := os.Getenv("OPENAGENDA_API_URL"); value != "" {
		config.Ingestion.OpenAgendaAPIURL = value
	}
	if value := os.Getenv("OPENAGENDA_API_KEY"); value != "" {
		config.Ingestion.OpenAgendaAPIKey = value
	}
	if value := os.Getenv("OPENAGENDA_OFFICIAL_ONLY"); value != "" {
		if v, err := parseBool(value); err == nil {
			config.Ingestion.OpenAgendaOfficialOnly = v
		}
	}
	if value := os.Getenv("OPENAGENDA_AGENDA_UPDATED_WITHIN_DAYS"); value != "" {
		if v, err := parseInt(value); err == nil {
			config.Ingestion.OpenAgendaAgendaUpdatedWithinDays = v
		}
	}
	if value := os.Getenv("OPENAGENDA_MAX_AGENDAS"); value != "" {
		if v, err := parseInt(value); err == nil {
			config.Ingestion.OpenAgendaMaxAgendas = v
		}
	}
	if value := os.Getenv("HTTP_TIMEOUT_SECONDS"); value != "" {
		if v, err := parseInt(value); err == nil {
			config.Ingestion.HTTPTimeoutSeconds = v
		}
	}
	if value := os.Getenv("BAN_RETRY_ATTEMPTS"); value != "" {
		if v, err := parseInt(value); err == nil {
			config.Ingestion.BANRetryAttempts = v
		}
	}
	if value := os.Getenv("INGESTION_DEBUG_DIR"); value != "" {
		config.Ingestion.DebugDir = value
	}
	if value := os.Getenv("SCHEDULER_TIMEZONE"); value != "" {
		config.Ingestion.SchedulerTimezone = value
	}
	if value := os.Getenv("DATAGOUV_SEARCH_QUERIES"); value != "" {
		config.Ingestion.DataGouvQueries = splitCSV(value)
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
		Ingestion: IngestionConfig{
			BANAPIURL:                         "https://api-adresse.data.gouv.fr",
			DataGouvAPIURL:                    "https://www.data.gouv.fr/api/1",
			OpenAgendaAPIURL:                  "https://api.openagenda.com/v2",
			OpenAgendaOfficialOnly:            true,
			OpenAgendaAgendaUpdatedWithinDays: 30,
			HTTPTimeoutSeconds:                20,
			BANRetryAttempts:                  3,
			DebugDir:                          "data/ingestion-debug",
			SchedulerTimezone:                 "Europe/Paris",
			DataGouvQueries:                   []string{"agenda", "evenement", "culture", "festival", "mairie"},
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

func splitCSV(value string) []string {
	parts := strings.Split(value, ",")
	out := make([]string, 0, len(parts))
	for _, part := range parts {
		part = strings.TrimSpace(part)
		if part != "" {
			out = append(out, part)
		}
	}
	return out
}
