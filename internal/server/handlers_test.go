package server

import (
	"encoding/json"
	"log/slog"
	"net/http"
	"net/http/httptest"
	"os"
	"testing"

	"github.com/geekxflood/program-director/internal/config"
)

func TestWriteJSON(t *testing.T) {
	recorder := httptest.NewRecorder()

	data := map[string]string{"key": "value"}
	writeJSON(recorder, http.StatusOK, data)

	if recorder.Code != http.StatusOK {
		t.Errorf("expected status 200, got %d", recorder.Code)
	}

	contentType := recorder.Header().Get("Content-Type")
	if contentType != "application/json" {
		t.Errorf("expected Content-Type application/json, got %s", contentType)
	}

	var result map[string]string
	if err := json.NewDecoder(recorder.Body).Decode(&result); err != nil {
		t.Fatalf("failed to decode response: %v", err)
	}

	if result["key"] != "value" {
		t.Errorf("expected value 'value', got %s", result["key"])
	}
}

func TestHandleHealth(t *testing.T) {
	cfg := &config.Config{}
	serverCfg := &Config{Port: 8080, MetricsEnabled: true}
	logger := slog.New(slog.NewTextHandler(os.Stdout, nil))

	server := NewServer(cfg, serverCfg, nil, nil, nil, nil, nil, nil, logger)

	req := httptest.NewRequest(http.MethodGet, "/health", nil)
	recorder := httptest.NewRecorder()

	server.handleHealth(recorder, req)

	if recorder.Code != http.StatusOK {
		t.Errorf("expected status 200, got %d", recorder.Code)
	}

	var result map[string]interface{}
	if err := json.NewDecoder(recorder.Body).Decode(&result); err != nil {
		t.Fatalf("failed to decode response: %v", err)
	}

	if result["status"] != "healthy" {
		t.Errorf("expected status healthy, got %v", result["status"])
	}

	if _, ok := result["timestamp"]; !ok {
		t.Error("expected timestamp in response")
	}

	if _, ok := result["version"]; !ok {
		t.Error("expected version in response")
	}
}

func TestHandleHealthMethodNotAllowed(t *testing.T) {
	cfg := &config.Config{}
	serverCfg := &Config{Port: 8080, MetricsEnabled: true}
	logger := slog.New(slog.NewTextHandler(os.Stdout, nil))

	server := NewServer(cfg, serverCfg, nil, nil, nil, nil, nil, nil, logger)

	req := httptest.NewRequest(http.MethodPost, "/health", nil)
	recorder := httptest.NewRecorder()

	server.handleHealth(recorder, req)

	if recorder.Code != http.StatusMethodNotAllowed {
		t.Errorf("expected status 405, got %d", recorder.Code)
	}
}

func TestHandleThemesList(t *testing.T) {
	cfg := &config.Config{
		Themes: []config.ThemeConfig{
			{Name: "theme1", ChannelID: "ch1"},
			{Name: "theme2", ChannelID: "ch2"},
		},
	}
	serverCfg := &Config{Port: 8080, MetricsEnabled: true}
	logger := slog.New(slog.NewTextHandler(os.Stdout, nil))

	server := NewServer(cfg, serverCfg, nil, nil, nil, nil, nil, nil, logger)

	req := httptest.NewRequest(http.MethodGet, "/api/v1/themes", nil)
	recorder := httptest.NewRecorder()

	server.handleThemesList(recorder, req)

	if recorder.Code != http.StatusOK {
		t.Errorf("expected status 200, got %d", recorder.Code)
	}

	var result successResponse
	if err := json.NewDecoder(recorder.Body).Decode(&result); err != nil {
		t.Fatalf("failed to decode response: %v", err)
	}

	if !result.Success {
		t.Error("expected success to be true")
	}

	data, ok := result.Data.(map[string]interface{})
	if !ok {
		t.Fatal("expected data to be map")
	}

	count, ok := data["count"].(float64)
	if !ok || int(count) != 2 {
		t.Errorf("expected count 2, got %v", data["count"])
	}
}

func TestHandleMetrics(t *testing.T) {
	// Skip this test as it requires database mocking
	// which is complex. The metrics endpoint is tested
	// through integration tests.
	t.Skip("Skipping metrics test - requires database mocking")
}

func TestServerNew(t *testing.T) {
	cfg := &config.Config{}
	serverCfg := &Config{Port: 8080, MetricsEnabled: true}
	logger := slog.New(slog.NewTextHandler(os.Stdout, nil))

	server := NewServer(cfg, serverCfg, nil, nil, nil, nil, nil, nil, logger)

	if server == nil {
		t.Fatal("expected non-nil server")
	}

	if server.config != cfg {
		t.Error("expected config to be set")
	}

	if server.logger != logger {
		t.Error("expected logger to be set")
	}

	if server.metricsEnabled != true {
		t.Error("expected metrics to be enabled")
	}
}

func contains(s, substr string) bool {
	return len(s) >= len(substr) && (s == substr || len(s) > len(substr) &&
		(s[:len(substr)] == substr || contains(s[1:], substr)))
}
