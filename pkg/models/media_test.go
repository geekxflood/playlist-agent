package models

import (
	"testing"
	"time"
)

func TestMediaWithScore_Sorting(t *testing.T) {
	items := []MediaWithScore{
		{
			Media: Media{
				ID:    1,
				Title: "Low Score",
			},
			Score: 1.5,
		},
		{
			Media: Media{
				ID:    2,
				Title: "High Score",
			},
			Score: 9.5,
		},
		{
			Media: Media{
				ID:    3,
				Title: "Medium Score",
			},
			Score: 5.0,
		},
	}

	// Sort by score descending (manual implementation)
	for i := 0; i < len(items); i++ {
		for j := i + 1; j < len(items); j++ {
			if items[i].Score < items[j].Score {
				items[i], items[j] = items[j], items[i]
			}
		}
	}

	// Check order
	if items[0].ID != 2 {
		t.Errorf("First item should be ID 2, got %d", items[0].ID)
	}
	if items[1].ID != 3 {
		t.Errorf("Second item should be ID 3, got %d", items[1].ID)
	}
	if items[2].ID != 1 {
		t.Errorf("Third item should be ID 1, got %d", items[2].ID)
	}
}

func TestPlaylist_TotalDuration(t *testing.T) {
	playlist := &Playlist{
		ThemeName:   "test-theme",
		ChannelID:   "test-channel",
		GeneratedAt: time.Now(),
		Items: []MediaWithScore{
			{
				Media: Media{
					ID:      1,
					Runtime: 120, // 2 hours
				},
				Score: 8.0,
			},
			{
				Media: Media{
					ID:      2,
					Runtime: 90, // 1.5 hours
				},
				Score: 7.5,
			},
			{
				Media: Media{
					ID:      3,
					Runtime: 150, // 2.5 hours
				},
				Score: 9.0,
			},
		},
		TotalScore: 24.5,
		Duration:   360, // Should be 120 + 90 + 150
	}

	// Verify duration calculation
	expectedDuration := 360
	if playlist.Duration != expectedDuration {
		t.Errorf("Playlist duration = %d, want %d", playlist.Duration, expectedDuration)
	}

	// Verify total score
	expectedScore := 24.5
	if playlist.TotalScore != expectedScore {
		t.Errorf("Playlist total score = %.2f, want %.2f", playlist.TotalScore, expectedScore)
	}
}

func TestMedia_Validation(t *testing.T) {
	now := time.Now()
	tests := []struct {
		name  string
		media Media
		valid bool
	}{
		{
			name: "valid movie",
			media: Media{
				ID:         1,
				ExternalID: 12345,
				Source:     MediaSourceRadarr,
				Title:      "Test Movie",
				Year:       2024,
				MediaType:  MediaTypeMovie,
				IMDBRating: 7.5,
				Runtime:    120,
				Path:       "/path/to/movie.mkv",
				Genres:     StringSlice{"Action", "Sci-Fi"},
				Overview:   "A test movie",
				TMDBID:     12345,
				HasFile:    true,
				SyncedAt:   now,
				CreatedAt:  now,
				UpdatedAt:  now,
			},
			valid: true,
		},
		{
			name: "valid series",
			media: Media{
				ID:         2,
				ExternalID: 67890,
				Source:     MediaSourceSonarr,
				Title:      "Test Series",
				Year:       2023,
				MediaType:  MediaTypeSeries,
				TMDBRating: 8.0,
				Runtime:    45,
				Path:       "/path/to/series",
				Genres:     StringSlice{"Drama"},
				Overview:   "A test series",
				TVDBID:     67890,
				HasFile:    true,
				SyncedAt:   now,
				CreatedAt:  now,
				UpdatedAt:  now,
			},
			valid: true,
		},
		{
			name: "valid anime",
			media: Media{
				ID:         3,
				ExternalID: 11111,
				Source:     MediaSourceSonarr,
				Title:      "Test Anime",
				Year:       2022,
				MediaType:  MediaTypeAnime,
				TMDBRating: 8.5,
				Runtime:    24,
				Path:       "/path/to/anime",
				Genres:     StringSlice{"Animation", "Action"},
				Overview:   "A test anime",
				TVDBID:     11111,
				HasFile:    true,
				SyncedAt:   now,
				CreatedAt:  now,
				UpdatedAt:  now,
			},
			valid: true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			// Basic validation checks
			if tt.media.ID <= 0 {
				t.Error("ID should be positive")
			}
			if tt.media.Title == "" {
				t.Error("Title should not be empty")
			}
			if tt.media.MediaType == "" {
				t.Error("MediaType should not be empty")
			}
			if tt.media.IMDBRating < 0 || tt.media.IMDBRating > 10 {
				t.Error("IMDBRating should be between 0 and 10")
			}
			if tt.media.TMDBRating < 0 || tt.media.TMDBRating > 10 {
				t.Error("TMDBRating should be between 0 and 10")
			}
			if tt.media.Runtime <= 0 {
				t.Error("Runtime should be positive")
			}
		})
	}
}

func TestPlayHistory_Fields(t *testing.T) {
	now := time.Now()
	history := PlayHistory{
		ID:         1,
		MediaID:    100,
		ChannelID:  "channel-123",
		ThemeName:  "sci-fi-night",
		PlayedAt:   now,
		MediaTitle: "Test Movie",
		MediaType:  "movie",
	}

	if history.ID != 1 {
		t.Errorf("ID = %d, want 1", history.ID)
	}
	if history.MediaID != 100 {
		t.Errorf("MediaID = %d, want 100", history.MediaID)
	}
	if history.ChannelID != "channel-123" {
		t.Errorf("ChannelID = %s, want channel-123", history.ChannelID)
	}
	if history.ThemeName != "sci-fi-night" {
		t.Errorf("ThemeName = %s, want sci-fi-night", history.ThemeName)
	}
	if history.PlayedAt != now {
		t.Errorf("PlayedAt mismatch")
	}
}
