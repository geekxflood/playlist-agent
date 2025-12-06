// Package trakt provides a client for interacting with the Trakt.tv API.
package trakt

import (
	"context"
	"encoding/json"
	"fmt"
	"net/http"
	"time"

	"github.com/geekxflood/program-director/internal/config"
)

const (
	baseURL        = "https://api.trakt.tv"
	apiVersion     = "2"
	defaultTimeout = 30 * time.Second
)

// Client represents a Trakt.tv API client
type Client struct {
	baseURL    string
	clientID   string
	httpClient *http.Client
}

// New creates a new Trakt client
func New(cfg *config.TraktConfig) *Client {
	return &Client{
		baseURL:  baseURL,
		clientID: cfg.ClientID,
		httpClient: &http.Client{
			Timeout: defaultTimeout,
		},
	}
}

// doRequest performs an HTTP request with Trakt API headers
func (c *Client) doRequest(ctx context.Context, method, path string, result interface{}) error {
	url := fmt.Sprintf("%s%s", c.baseURL, path)

	req, err := http.NewRequestWithContext(ctx, method, url, nil)
	if err != nil {
		return fmt.Errorf("failed to create request: %w", err)
	}

	// Set Trakt API headers
	req.Header.Set("Content-Type", "application/json")
	req.Header.Set("trakt-api-version", apiVersion)
	req.Header.Set("trakt-api-key", c.clientID)

	resp, err := c.httpClient.Do(req)
	if err != nil {
		return fmt.Errorf("request failed: %w", err)
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		return fmt.Errorf("unexpected status code: %d", resp.StatusCode)
	}

	if result != nil {
		if err := json.NewDecoder(resp.Body).Decode(result); err != nil {
			return fmt.Errorf("failed to decode response: %w", err)
		}
	}

	return nil
}

// Movie represents a Trakt movie
type Movie struct {
	Title string `json:"title"`
	Year  int    `json:"year"`
	IDs   struct {
		Trakt int    `json:"trakt"`
		Slug  string `json:"slug"`
		IMDB  string `json:"imdb"`
		TMDB  int    `json:"tmdb"`
	} `json:"ids"`
	Tagline  string   `json:"tagline"`
	Overview string   `json:"overview"`
	Released string   `json:"released"`
	Runtime  int      `json:"runtime"` // minutes
	Genres   []string `json:"genres"`
	Rating   float64  `json:"rating"`
	Votes    int      `json:"votes"`
}

// Show represents a Trakt TV show
type Show struct {
	Title string `json:"title"`
	Year  int    `json:"year"`
	IDs   struct {
		Trakt int    `json:"trakt"`
		Slug  string `json:"slug"`
		IMDB  string `json:"imdb"`
		TMDB  int    `json:"tmdb"`
		TVDB  int    `json:"tvdb"`
	} `json:"ids"`
	Overview string   `json:"overview"`
	Genres   []string `json:"genres"`
	Status   string   `json:"status"`
	Rating   float64  `json:"rating"`
	Votes    int      `json:"votes"`
	Network  string   `json:"network"`
	Country  string   `json:"country"`
}

// TrendingMovie represents a trending movie with stats
type TrendingMovie struct {
	Watchers int    `json:"watchers"`
	Movie    *Movie `json:"movie"`
}

// TrendingShow represents a trending show with stats
type TrendingShow struct {
	Watchers int   `json:"watchers"`
	Show     *Show `json:"show"`
}

// PopularMovie represents a popular movie
type PopularMovie struct {
	Movie *Movie `json:"movie"`
}

// PopularShow represents a popular show
type PopularShow struct {
	Show *Show `json:"show"`
}

// SearchResult represents a search result
type SearchResult struct {
	Type  string `json:"type"`
	Score float64 `json:"score"`
	Movie *Movie `json:"movie,omitempty"`
	Show  *Show  `json:"show,omitempty"`
}

// GetTrendingMovies retrieves currently trending movies
func (c *Client) GetTrendingMovies(ctx context.Context, limit int) ([]TrendingMovie, error) {
	if limit == 0 {
		limit = 10
	}
	var movies []TrendingMovie
	path := fmt.Sprintf("/movies/trending?extended=full&limit=%d", limit)
	if err := c.doRequest(ctx, http.MethodGet, path, &movies); err != nil {
		return nil, err
	}
	return movies, nil
}

// GetTrendingShows retrieves currently trending shows
func (c *Client) GetTrendingShows(ctx context.Context, limit int) ([]TrendingShow, error) {
	if limit == 0 {
		limit = 10
	}
	var shows []TrendingShow
	path := fmt.Sprintf("/shows/trending?extended=full&limit=%d", limit)
	if err := c.doRequest(ctx, http.MethodGet, path, &shows); err != nil {
		return nil, err
	}
	return shows, nil
}

// GetPopularMovies retrieves popular movies
func (c *Client) GetPopularMovies(ctx context.Context, limit int) ([]Movie, error) {
	if limit == 0 {
		limit = 10
	}
	var movies []Movie
	path := fmt.Sprintf("/movies/popular?extended=full&limit=%d", limit)
	if err := c.doRequest(ctx, http.MethodGet, path, &movies); err != nil {
		return nil, err
	}
	return movies, nil
}

// GetPopularShows retrieves popular shows
func (c *Client) GetPopularShows(ctx context.Context, limit int) ([]Show, error) {
	if limit == 0 {
		limit = 10
	}
	var shows []Show
	path := fmt.Sprintf("/shows/popular?extended=full&limit=%d", limit)
	if err := c.doRequest(ctx, http.MethodGet, path, &shows); err != nil {
		return nil, err
	}
	return shows, nil
}

// Search performs a general search across movies and shows
func (c *Client) Search(ctx context.Context, query string, limit int) ([]SearchResult, error) {
	if limit == 0 {
		limit = 10
	}
	var results []SearchResult
	path := fmt.Sprintf("/search/movie,show?query=%s&extended=full&limit=%d", query, limit)
	if err := c.doRequest(ctx, http.MethodGet, path, &results); err != nil {
		return nil, err
	}
	return results, nil
}
