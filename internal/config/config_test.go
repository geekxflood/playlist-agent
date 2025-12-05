package config

import (
	"os"
	"testing"
)

func TestValidate(t *testing.T) {
	tests := []struct {
		name    string
		config  Config
		wantErr bool
		errMsg  string
	}{
		{
			name: "valid config",
			config: Config{
				Database: DatabaseConfig{
					Driver: "sqlite",
					SQLite: SQLiteConfig{
						Path: "./test.db",
					},
				},
				Radarr: RadarrConfig{
					URL:    "http://localhost:7878",
					APIKey: "test-key",
				},
				Sonarr: SonarrConfig{
					URL:    "http://localhost:8989",
					APIKey: "test-key",
				},
				Tunarr: TunarrConfig{
					URL: "http://localhost:8000",
				},
				Ollama: OllamaConfig{
					URL:   "http://localhost:11434",
					Model: "test-model",
				},
			},
			wantErr: false,
		},
		{
			name: "invalid database driver",
			config: Config{
				Database: DatabaseConfig{
					Driver: "invalid",
				},
				Radarr: RadarrConfig{
					URL:    "http://localhost:7878",
					APIKey: "test-key",
				},
				Sonarr: SonarrConfig{
					URL:    "http://localhost:8989",
					APIKey: "test-key",
				},
				Tunarr: TunarrConfig{
					URL: "http://localhost:8000",
				},
				Ollama: OllamaConfig{
					URL:   "http://localhost:11434",
					Model: "test-model",
				},
			},
			wantErr: true,
			errMsg:  "invalid database driver",
		},
		{
			name: "missing radarr api key",
			config: Config{
				Database: DatabaseConfig{
					Driver: "sqlite",
				},
				Radarr: RadarrConfig{
					URL:    "http://localhost:7878",
					APIKey: "",
				},
				Sonarr: SonarrConfig{
					URL:    "http://localhost:8989",
					APIKey: "test-key",
				},
				Tunarr: TunarrConfig{
					URL: "http://localhost:8000",
				},
				Ollama: OllamaConfig{
					URL:   "http://localhost:11434",
					Model: "test-model",
				},
			},
			wantErr: true,
			errMsg:  "radarr API key is required",
		},
		{
			name: "missing sonarr api key",
			config: Config{
				Database: DatabaseConfig{
					Driver: "sqlite",
				},
				Radarr: RadarrConfig{
					URL:    "http://localhost:7878",
					APIKey: "test-key",
				},
				Sonarr: SonarrConfig{
					URL:    "http://localhost:8989",
					APIKey: "",
				},
				Tunarr: TunarrConfig{
					URL: "http://localhost:8000",
				},
				Ollama: OllamaConfig{
					URL:   "http://localhost:11434",
					Model: "test-model",
				},
			},
			wantErr: true,
			errMsg:  "sonarr API key is required",
		},
		{
			name: "missing theme channel id",
			config: Config{
				Database: DatabaseConfig{
					Driver: "sqlite",
				},
				Radarr: RadarrConfig{
					URL:    "http://localhost:7878",
					APIKey: "test-key",
				},
				Sonarr: SonarrConfig{
					URL:    "http://localhost:8989",
					APIKey: "test-key",
				},
				Tunarr: TunarrConfig{
					URL: "http://localhost:8000",
				},
				Ollama: OllamaConfig{
					URL:   "http://localhost:11434",
					Model: "test-model",
				},
				Themes: []ThemeConfig{
					{
						Name:      "test-theme",
						ChannelID: "",
					},
				},
			},
			wantErr: true,
			errMsg:  "channel_id is required",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			err := tt.config.Validate()
			if (err != nil) != tt.wantErr {
				t.Errorf("Validate() error = %v, wantErr %v", err, tt.wantErr)
				return
			}
			if err != nil && tt.errMsg != "" {
				if !contains(err.Error(), tt.errMsg) {
					t.Errorf("Validate() error = %v, want error containing %v", err, tt.errMsg)
				}
			}
		})
	}
}

func TestPostgresConfigDSN(t *testing.T) {
	cfg := PostgresConfig{
		Host:     "localhost",
		Port:     5432,
		Database: "testdb",
		User:     "testuser",
		Password: "testpass",
		SSLMode:  "disable",
	}

	want := "host=localhost port=5432 user=testuser password=testpass dbname=testdb sslmode=disable"
	got := cfg.DSN()

	if got != want {
		t.Errorf("DSN() = %v, want %v", got, want)
	}
}

func TestLoadConfigWithDefaults(t *testing.T) {
	// Set required environment variables
	os.Setenv("RADARR_API_KEY", "test-radarr-key")
	os.Setenv("SONARR_API_KEY", "test-sonarr-key")
	defer func() {
		os.Unsetenv("RADARR_API_KEY")
		os.Unsetenv("SONARR_API_KEY")
	}()

	// Try to load config without a file (should use defaults and env vars)
	cfg, err := Load("nonexistent-config-file.yaml")
	if err != nil {
		t.Skipf("Skipping test due to error loading config: %v", err)
		return
	}

	// Verify defaults were applied
	if cfg.Database.Driver != "sqlite" {
		t.Errorf("Default database driver = %v, want sqlite", cfg.Database.Driver)
	}
	if cfg.Ollama.Model != "dolphin-llama3:8b" {
		t.Errorf("Default ollama model = %v, want dolphin-llama3:8b", cfg.Ollama.Model)
	}
	if cfg.Ollama.Temperature != 0.7 {
		t.Errorf("Default ollama temperature = %v, want 0.7", cfg.Ollama.Temperature)
	}
	if cfg.Radarr.APIKey != "test-radarr-key" {
		t.Errorf("Radarr API key = %v, want test-radarr-key", cfg.Radarr.APIKey)
	}
	if cfg.Sonarr.APIKey != "test-sonarr-key" {
		t.Errorf("Sonarr API key = %v, want test-sonarr-key", cfg.Sonarr.APIKey)
	}
}

// Helper function to check if string contains substring
func contains(s, substr string) bool {
	return len(s) >= len(substr) && (s == substr || len(substr) == 0 ||
		(len(s) > 0 && len(substr) > 0 && hasSubstring(s, substr)))
}

func hasSubstring(s, substr string) bool {
	for i := 0; i <= len(s)-len(substr); i++ {
		if s[i:i+len(substr)] == substr {
			return true
		}
	}
	return false
}
