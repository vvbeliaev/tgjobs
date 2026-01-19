package config

import (
	"os"
	"strconv"
)

// Config holds all application configuration.
type Config struct {
	Telegram TelegramConfig
	OpenAI   OpenAIConfig
}

// TelegramConfig holds Telegram API credentials.
type TelegramConfig struct {
	APIID       int
	APIHash     string
	Phone       string
	SessionPath string
}

// OpenAIConfig holds OpenAI API credentials.
type OpenAIConfig struct {
	APIKey  string
	BaseURL string
}

// Load reads configuration from environment variables.
func Load() Config {
	apiID, _ := strconv.Atoi(os.Getenv("TG_API_ID"))

	return Config{
		Telegram: TelegramConfig{
			APIID:       apiID,
			APIHash:     os.Getenv("TG_API_HASH"),
			Phone:       os.Getenv("TG_PHONE"),
			SessionPath: getEnvOrDefault("TG_SESSION_PATH", "session.json"),
		},
		OpenAI: OpenAIConfig{
			APIKey:  os.Getenv("OPENAI_API_KEY"),
			BaseURL: os.Getenv("OPENAI_BASE_URL"),
		},
	}
}

func getEnvOrDefault(key, defaultVal string) string {
	if v := os.Getenv(key); v != "" {
		return v
	}
	return defaultVal
}
