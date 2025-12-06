package trakt

import (
	"context"
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/geekxflood/program-director/internal/config"
)

func TestNew(t *testing.T) {
	cfg := &config.TraktConfig{
		ClientID: "test-client-id",
	}

	client := New(cfg)

	if client == nil {
		t.Fatal("expected non-nil client")
	}

	if client.baseURL != baseURL {
		t.Errorf("expected baseURL %s, got %s", baseURL, client.baseURL)
	}

	if client.clientID != cfg.ClientID {
		t.Errorf("expected clientID %s, got %s", cfg.ClientID, client.clientID)
	}

	if client.httpClient == nil {
		t.Error("expected non-nil httpClient")
	}
}

func TestGetTrendingMovies(t *testing.T) {
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Content-Type", "application/json")
		w.WriteHeader(http.StatusOK)
		w.Write([]byte(`[
			{
				"watchers": 5000,
				"movie": {
					"title": "Test Movie",
					"year": 2023,
					"ids": {"trakt": 1},
					"rating": 8.0
				}
			}
		]`))
	}))
	defer server.Close()

	cfg := &config.TraktConfig{ClientID: "test-key"}
	client := New(cfg)
	client.baseURL = server.URL

	movies, err := client.GetTrendingMovies(context.Background(), 10)
	if err != nil {
		t.Fatalf("expected no error, got %v", err)
	}

	if len(movies) != 1 {
		t.Errorf("expected 1 movie, got %d", len(movies))
	}

	if movies[0].Watchers != 5000 {
		t.Errorf("expected 5000 watchers, got %d", movies[0].Watchers)
	}

	if movies[0].Movie.Title != "Test Movie" {
		t.Errorf("expected title 'Test Movie', got %s", movies[0].Movie.Title)
	}
}

func TestSearch(t *testing.T) {
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		query := r.URL.Query().Get("query")
		if query == "" {
			t.Error("expected query parameter")
		}

		w.Header().Set("Content-Type", "application/json")
		w.WriteHeader(http.StatusOK)
		w.Write([]byte(`[
			{
				"type": "movie",
				"score": 100.0,
				"movie": {
					"title": "Searched Movie",
					"year": 2020,
					"ids": {"trakt": 1}
				}
			}
		]`))
	}))
	defer server.Close()

	cfg := &config.TraktConfig{ClientID: "test-key"}
	client := New(cfg)
	client.baseURL = server.URL

	results, err := client.Search(context.Background(), "test", 10)
	if err != nil {
		t.Fatalf("expected no error, got %v", err)
	}

	if len(results) != 1 {
		t.Errorf("expected 1 result, got %d", len(results))
	}

	if results[0].Type != "movie" {
		t.Errorf("expected type movie, got %s", results[0].Type)
	}

	if results[0].Score != 100.0 {
		t.Errorf("expected score 100.0, got %.1f", results[0].Score)
	}
}

func TestDoRequestError(t *testing.T) {
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusInternalServerError)
	}))
	defer server.Close()

	cfg := &config.TraktConfig{ClientID: "test-key"}
	client := New(cfg)
	client.baseURL = server.URL

	_, err := client.Search(context.Background(), "test", 10)
	if err == nil {
		t.Error("expected error for 500 response, got nil")
	}
}
