package config

import (
	"fmt"
	"os"
	"path/filepath"

	"github.com/joho/godotenv"
	"github.com/spf13/viper"
)

// Config holds all configuration for the application
type Config struct {
	PlaneBaseURL    string
	PlaneAPIToken   string
	PlaneWorkspace  string
	DefaultProject  string
	RequestTimeout  int
	TemplatesDir    string
	FuzzyMinScore   int
	FuzzyMaxResults int
}

// Load loads configuration from environment and config file
func Load() (*Config, error) {
	// Load .env file if exists
	envFile := ".env"
	if _, err := os.Stat(envFile); err == nil {
		if err := godotenv.Load(envFile); err != nil {
			return nil, fmt.Errorf("failed to load .env file: %w", err)
		}
	}

	// Initialize viper
	viper.SetConfigName("config")
	viper.SetConfigType("yaml")
	viper.AddConfigPath(".")
	viper.AddConfigPath("$HOME/.plane-cli")

	// Set defaults
	viper.SetDefault("defaults.project", "")
	viper.SetDefault("defaults.state", "Backlog")
	viper.SetDefault("defaults.priority", 2) // Medium
	viper.SetDefault("templates.directory", "./templates")
	viper.SetDefault("templates.default", "feature")
	viper.SetDefault("fuzzy.min_score", 60)
	viper.SetDefault("fuzzy.max_results", 10)
	viper.SetDefault("request.timeout", 30)

	// Read config file (optional)
	if err := viper.ReadInConfig(); err != nil {
		if _, ok := err.(viper.ConfigFileNotFoundError); !ok {
			return nil, fmt.Errorf("failed to read config file: %w", err)
		}
	}

	// Build config
	cfg := &Config{
		PlaneBaseURL:    getEnvOrDefault("PLANE_BASE_URL", ""),
		PlaneAPIToken:   getEnvOrDefault("PLANE_API_TOKEN", ""),
		PlaneWorkspace:  getEnvOrDefault("PLANE_WORKSPACE", ""),
		DefaultProject:  viper.GetString("defaults.project"),
		RequestTimeout:  viper.GetInt("request.timeout"),
		TemplatesDir:    viper.GetString("templates.directory"),
		FuzzyMinScore:   viper.GetInt("fuzzy.min_score"),
		FuzzyMaxResults: viper.GetInt("fuzzy.max_results"),
	}

	// Validate required fields
	if cfg.PlaneBaseURL == "" {
		return nil, fmt.Errorf("PLANE_BASE_URL is required")
	}
	if cfg.PlaneAPIToken == "" {
		return nil, fmt.Errorf("PLANE_API_TOKEN is required")
	}

	// Resolve templates directory
	if !filepath.IsAbs(cfg.TemplatesDir) {
		absPath, err := filepath.Abs(cfg.TemplatesDir)
		if err != nil {
			return nil, fmt.Errorf("failed to resolve templates directory: %w", err)
		}
		cfg.TemplatesDir = absPath
	}

	return cfg, nil
}

func getEnvOrDefault(key, defaultValue string) string {
	if value := os.Getenv(key); value != "" {
		return value
	}
	return defaultValue
}

// Validate checks if the configuration is valid
func (c *Config) Validate() error {
	if c.PlaneBaseURL == "" {
		return fmt.Errorf("Plane base URL is required")
	}
	if c.PlaneAPIToken == "" {
		return fmt.Errorf("Plane API token is required")
	}
	return nil
}
