// Package config provides configuration management for program-director.
package config

import (
	"errors"
	"fmt"
	"os"
	"path/filepath"
	"strings"

	"github.com/spf13/viper"
)

// Config holds all application configuration
type Config struct {
	Debug    bool           `mapstructure:"debug"`
	Database DatabaseConfig `mapstructure:"database"`
	Radarr   RadarrConfig   `mapstructure:"radarr"`
	Sonarr   SonarrConfig   `mapstructure:"sonarr"`
	Tunarr   TunarrConfig   `mapstructure:"tunarr"`
	Trakt    TraktConfig    `mapstructure:"trakt"`
	Ollama   OllamaConfig   `mapstructure:"ollama"`
	Cooldown CooldownConfig `mapstructure:"cooldown"`
	Server   ServerConfig   `mapstructure:"server"`
	Themes   []ThemeConfig  `mapstructure:"themes"`
}

// DatabaseConfig configures the database connection
type DatabaseConfig struct {
	Driver   string         `mapstructure:"driver"` // postgres or sqlite
	Postgres PostgresConfig `mapstructure:"postgres"`
	SQLite   SQLiteConfig   `mapstructure:"sqlite"`
}

// PostgresConfig holds PostgreSQL connection settings
type PostgresConfig struct {
	Host     string `mapstructure:"host"`
	Port     int    `mapstructure:"port"`
	Database string `mapstructure:"database"`
	User     string `mapstructure:"user"`
	Password string `mapstructure:"password"`
	SSLMode  string `mapstructure:"sslmode"`
}

// SQLiteConfig holds SQLite settings
type SQLiteConfig struct {
	Path string `mapstructure:"path"`
}

// RadarrConfig holds Radarr API settings
type RadarrConfig struct {
	URL    string `mapstructure:"url"`
	APIKey string `mapstructure:"api_key"`
}

// SonarrConfig holds Sonarr API settings
type SonarrConfig struct {
	URL    string `mapstructure:"url"`
	APIKey string `mapstructure:"api_key"`
}

// TunarrConfig holds Tunarr API settings
type TunarrConfig struct {
	URL string `mapstructure:"url"`
}

// TraktConfig holds Trakt.tv API settings
type TraktConfig struct {
	ClientID     string `mapstructure:"client_id"`
	ClientSecret string `mapstructure:"client_secret"`
}

// OllamaConfig holds Ollama LLM settings
type OllamaConfig struct {
	URL         string  `mapstructure:"url"`
	Model       string  `mapstructure:"model"`
	Temperature float64 `mapstructure:"temperature"`
	NumCtx      int     `mapstructure:"num_ctx"`
}

// CooldownConfig holds media cooldown settings
type CooldownConfig struct {
	MovieDays  int `mapstructure:"movie_days"`
	SeriesDays int `mapstructure:"series_days"`
	AnimeDays  int `mapstructure:"anime_days"`
}

// ServerConfig holds HTTP server settings
type ServerConfig struct {
	Port            int  `mapstructure:"port"`
	EnableScheduler bool `mapstructure:"enable_scheduler"`
	MetricsEnabled  bool `mapstructure:"metrics_enabled"`
	ShutdownTimeout int  `mapstructure:"shutdown_timeout"`
}

// ThemeConfig defines a playlist theme
type ThemeConfig struct {
	Name        string   `mapstructure:"name"`
	Description string   `mapstructure:"description"`
	ChannelID   string   `mapstructure:"channel_id"`
	Schedule    string   `mapstructure:"schedule"`
	MediaTypes  []string `mapstructure:"media_types"`
	Genres      []string `mapstructure:"genres"`
	Keywords    []string `mapstructure:"keywords"`
	MinRating   float64  `mapstructure:"min_rating"`
	MaxItems    int      `mapstructure:"max_items"`
	Duration    int      `mapstructure:"duration"` // Target duration in minutes
}

// Load reads configuration from file and environment variables
func Load(configFile string) (*Config, error) {
	v := viper.New()

	// Set defaults
	setDefaults(v)

	// Determine config file path
	if configFile != "" {
		v.SetConfigFile(configFile)
	} else {
		// Search for config in common locations
		v.SetConfigName("config")
		v.SetConfigType("yaml")
		v.AddConfigPath(".")
		v.AddConfigPath("./configs")
		v.AddConfigPath("/etc/program-director")

		// Also check home directory
		if home, err := os.UserHomeDir(); err == nil {
			v.AddConfigPath(filepath.Join(home, ".config", "program-director"))
		}
	}

	// Read config file
	if err := v.ReadInConfig(); err != nil {
		var configFileNotFoundError viper.ConfigFileNotFoundError
		if errors.As(err, &configFileNotFoundError) {
			return nil, fmt.Errorf("error reading config file: %w", err)
		}
		// Config file not found is okay, we'll use defaults and env vars
	}

	// Environment variable overrides
	v.SetEnvPrefix("PROGRAMDIR")
	v.SetEnvKeyReplacer(strings.NewReplacer(".", "_"))
	v.AutomaticEnv()

	// Map specific environment variables
	bindEnvVars(v)

	var cfg Config
	if err := v.Unmarshal(&cfg); err != nil {
		return nil, fmt.Errorf("error unmarshaling config: %w", err)
	}

	// Validate configuration
	if err := cfg.Validate(); err != nil {
		return nil, fmt.Errorf("config validation error: %w", err)
	}

	return &cfg, nil
}

// setDefaults configures default values
func setDefaults(v *viper.Viper) {
	// Database defaults
	v.SetDefault("database.driver", "sqlite")
	v.SetDefault("database.postgres.host", "localhost")
	v.SetDefault("database.postgres.port", 5432)
	v.SetDefault("database.postgres.database", "program_director")
	v.SetDefault("database.postgres.sslmode", "disable")
	v.SetDefault("database.sqlite.path", "./data/program-director.db")

	// Radarr defaults
	v.SetDefault("radarr.url", "http://radarr:7878")

	// Sonarr defaults
	v.SetDefault("sonarr.url", "http://sonarr:8989")

	// Tunarr defaults
	v.SetDefault("tunarr.url", "http://tunarr:8000")

	// Trakt defaults (optional, no defaults needed)

	// Ollama defaults
	v.SetDefault("ollama.url", "http://ollama:11434")
	v.SetDefault("ollama.model", "dolphin-llama3:8b")
	v.SetDefault("ollama.temperature", 0.7)
	v.SetDefault("ollama.num_ctx", 8192)

	// Cooldown defaults
	v.SetDefault("cooldown.movie_days", 30)
	v.SetDefault("cooldown.series_days", 14)
	v.SetDefault("cooldown.anime_days", 14)

	// Server defaults
	v.SetDefault("server.port", 8080)
	v.SetDefault("server.enable_scheduler", false)
	v.SetDefault("server.metrics_enabled", true)
	v.SetDefault("server.shutdown_timeout", 30)
}

// bindEnvVars maps environment variables to config keys
func bindEnvVars(v *viper.Viper) {
	// Direct environment variable mappings
	bindings := []struct {
		key string
		env string
	}{
		{"radarr.api_key", "RADARR_API_KEY"},
		{"sonarr.api_key", "SONARR_API_KEY"},
		{"radarr.url", "RADARR_URL"},
		{"sonarr.url", "SONARR_URL"},
		{"tunarr.url", "TUNARR_URL"},
		{"trakt.client_id", "TRAKT_CLIENT_ID"},
		{"trakt.client_secret", "TRAKT_CLIENT_SECRET"},
		{"ollama.url", "OLLAMA_URL"},
		{"ollama.model", "OLLAMA_MODEL"},
		{"database.driver", "DB_DRIVER"},
		{"database.postgres.host", "POSTGRES_HOST"},
		{"database.postgres.port", "POSTGRES_PORT"},
		{"database.postgres.database", "POSTGRES_DATABASE"},
		{"database.postgres.user", "POSTGRES_USER"},
		{"database.postgres.password", "POSTGRES_PASSWORD"},
	}

	for _, b := range bindings {
		if err := v.BindEnv(b.key, b.env); err != nil {
			panic(fmt.Sprintf("failed to bind env var %s: %v", b.env, err))
		}
	}
}

// Validate checks if the configuration is valid
func (c *Config) Validate() error {
	// Validate database config
	switch c.Database.Driver {
	case "postgres":
		if c.Database.Postgres.Host == "" {
			return errors.New("postgres host is required")
		}
	case "sqlite":
		// SQLite path can be empty (use default)
	default:
		return fmt.Errorf("invalid database driver: %s (must be postgres or sqlite)", c.Database.Driver)
	}

	// Validate Radarr config
	if c.Radarr.URL == "" {
		return errors.New("radarr URL is required")
	}
	if c.Radarr.APIKey == "" {
		return errors.New("radarr API key is required")
	}

	// Validate Sonarr config
	if c.Sonarr.URL == "" {
		return errors.New("sonarr URL is required")
	}
	if c.Sonarr.APIKey == "" {
		return errors.New("sonarr API key is required")
	}

	// Validate Tunarr config
	if c.Tunarr.URL == "" {
		return errors.New("tunarr URL is required")
	}

	// Validate Ollama config
	if c.Ollama.URL == "" {
		return errors.New("ollama URL is required")
	}
	if c.Ollama.Model == "" {
		return errors.New("ollama model is required")
	}

	// Validate themes
	for i, theme := range c.Themes {
		if theme.Name == "" {
			return fmt.Errorf("theme %d: name is required", i)
		}
		if theme.ChannelID == "" {
			return fmt.Errorf("theme %s: channel_id is required", theme.Name)
		}
	}

	return nil
}

// DSN returns the database connection string for PostgreSQL
func (c *PostgresConfig) DSN() string {
	return fmt.Sprintf(
		"host=%s port=%d user=%s password=%s dbname=%s sslmode=%s",
		c.Host, c.Port, c.User, c.Password, c.Database, c.SSLMode,
	)
}
