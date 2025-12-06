package cmd

import (
	"context"
	"fmt"
	"os"
	"os/signal"
	"syscall"

	"github.com/spf13/cobra"

	"github.com/geekxflood/program-director/internal/database"
	"github.com/geekxflood/program-director/internal/database/repository"
	"github.com/geekxflood/program-director/pkg/models"
)

var (
	scanDetailed bool
	scanSource   string
)

// scanCmd represents the scan command
var scanCmd = &cobra.Command{
	Use:   "scan",
	Short: "Display media library information",
	Long: `Scan and display information about the media library.

This command queries the local database and optionally the source
applications (Radarr/Sonarr) to display library statistics.

Examples:
  # Show library summary
  program-director scan

  # Show detailed information
  program-director scan --detailed

  # Scan specific source
  program-director scan --source radarr`,
	RunE: runScan,
}

func init() {
	scanCmd.Flags().BoolVarP(&scanDetailed, "detailed", "d", false, "show detailed information")
	scanCmd.Flags().StringVarP(&scanSource, "source", "s", "", "specific source to scan (radarr, sonarr)")
}

func runScan(_ *cobra.Command, _ []string) error {
	ctx, cancel := context.WithCancel(context.Background())
	defer cancel()

	// Handle graceful shutdown
	sigChan := make(chan os.Signal, 1)
	signal.Notify(sigChan, syscall.SIGINT, syscall.SIGTERM)
	go func() {
		<-sigChan
		logger.Info("received shutdown signal")
		cancel()
	}()

	logger.Info("scanning media library",
		"detailed", scanDetailed,
		"source", scanSource,
		"database_driver", cfg.Database.Driver,
	)

	logger.Debug("initializing database connection")
	db, err := database.New(ctx, &cfg.Database, logger)
	if err != nil {
		logger.Error("failed to initialize database", "error", err)
		return fmt.Errorf("failed to initialize database: %w", err)
	}
	defer func() {
		if err := db.Close(); err != nil {
			logger.Error("failed to close database", "error", err)
		}
	}()

	logger.Debug("database connection established")

	// Initialize repositories
	mediaRepo := repository.NewMediaRepository(db)
	historyRepo := repository.NewHistoryRepository(db)
	cooldownRepo := repository.NewCooldownRepository(db)

	logger.Debug("querying media statistics")

	// Get media counts by type
	stats, err := getMediaStatistics(ctx, mediaRepo, historyRepo, cooldownRepo, scanSource)
	if err != nil {
		logger.Error("failed to get statistics", "error", err)
		return fmt.Errorf("failed to get statistics: %w", err)
	}

	select {
	case <-ctx.Done():
		return ctx.Err()
	default:
	}

	logger.Debug("generating report")

	// Display results
	printMediaSummary(stats, scanDetailed)

	logger.Info("scan complete",
		"movies", stats.MovieCount,
		"series", stats.SeriesCount,
		"anime", stats.AnimeCount,
		"total_plays", stats.TotalPlays,
	)

	return nil
}

// MediaStatistics holds media library statistics
type MediaStatistics struct {
	MovieCount       int64
	SeriesCount      int64
	AnimeCount       int64
	TotalPlays       int64
	OnCooldown       int64
	TopGenres        map[string]int
	AverageRating    float64
	TotalSize        int64
	ConfiguredThemes int
}

// getMediaStatistics queries the database for media statistics
func getMediaStatistics(
	ctx context.Context,
	mediaRepo *repository.MediaRepository,
	historyRepo *repository.HistoryRepository,
	cooldownRepo *repository.CooldownRepository,
	source string,
) (*MediaStatistics, error) {
	stats := &MediaStatistics{
		TopGenres:        make(map[string]int),
		ConfiguredThemes: len(cfg.Themes),
	}

	// Count movies
	hasFile := true
	movieOpts := repository.ListMediaOptions{
		MediaType: models.MediaTypeMovie,
		HasFile:   &hasFile,
	}
	if source == "radarr" {
		movieOpts.Source = models.MediaSourceRadarr
	}
	var err error
	stats.MovieCount, err = mediaRepo.Count(ctx, movieOpts)
	if err != nil {
		return nil, fmt.Errorf("failed to count movies: %w", err)
	}

	// Count series
	seriesOpts := repository.ListMediaOptions{
		MediaType: models.MediaTypeSeries,
		HasFile:   &hasFile,
	}
	if source == "sonarr" {
		seriesOpts.Source = models.MediaSourceSonarr
	}
	stats.SeriesCount, err = mediaRepo.Count(ctx, seriesOpts)
	if err != nil {
		return nil, fmt.Errorf("failed to count series: %w", err)
	}

	// Count anime
	animeOpts := repository.ListMediaOptions{
		MediaType: models.MediaTypeAnime,
		HasFile:   &hasFile,
	}
	if source == "sonarr" {
		animeOpts.Source = models.MediaSourceSonarr
	}
	stats.AnimeCount, err = mediaRepo.Count(ctx, animeOpts)
	if err != nil {
		return nil, fmt.Errorf("failed to count anime: %w", err)
	}

	// Get play history count
	stats.TotalPlays, err = historyRepo.Count(ctx, repository.ListHistoryOptions{})
	if err != nil {
		return nil, fmt.Errorf("failed to count history: %w", err)
	}

	// Get cooldown count
	stats.OnCooldown, err = cooldownRepo.CountActive(ctx)
	if err != nil {
		return nil, fmt.Errorf("failed to count cooldowns: %w", err)
	}

	// Get all media for genre and rating stats
	allMedia, err := mediaRepo.List(ctx, repository.ListMediaOptions{
		HasFile: &hasFile,
		Limit:   1000,
	})
	if err == nil {
		totalRating := 0.0
		ratingCount := 0
		for _, m := range allMedia {
			// Count genres
			for _, genre := range m.Genres {
				stats.TopGenres[genre]++
			}

			// Average rating
			if m.IMDBRating > 0 {
				totalRating += m.IMDBRating
				ratingCount++
			} else if m.TMDBRating > 0 {
				totalRating += m.TMDBRating
				ratingCount++
			}

			// Total size
			stats.TotalSize += m.SizeOnDisk
		}
		if ratingCount > 0 {
			stats.AverageRating = totalRating / float64(ratingCount)
		}
	}

	return stats, nil
}

// printMediaSummary displays media statistics
func printMediaSummary(stats *MediaStatistics, detailed bool) {
	fmt.Println()
	fmt.Println("┌─────────────────────────────────────────────────────────┐")
	fmt.Println("│         Program Director - Media Library Summary        │")
	fmt.Println("└─────────────────────────────────────────────────────────┘")
	fmt.Println()

	// Configuration
	fmt.Printf("Configured Themes: %d\n", stats.ConfiguredThemes)
	if stats.ConfiguredThemes > 0 {
		fmt.Println("\nThemes:")
		for i, theme := range cfg.Themes {
			fmt.Printf("  %d. %-20s (Channel: %s)\n", i+1, theme.Name, theme.ChannelID)
		}
	}
	fmt.Println()

	// Media counts
	fmt.Println("Media Library")
	fmt.Println("─────────────")
	fmt.Printf("  Movies:     %6d\n", stats.MovieCount)
	fmt.Printf("  TV Shows:   %6d\n", stats.SeriesCount)
	fmt.Printf("  Anime:      %6d\n", stats.AnimeCount)
	fmt.Printf("  Total:      %6d\n", stats.MovieCount+stats.SeriesCount+stats.AnimeCount)
	fmt.Println()

	// Playback stats
	fmt.Println("Playback History")
	fmt.Println("────────────────")
	fmt.Printf("  Total plays:    %6d\n", stats.TotalPlays)
	fmt.Printf("  On cooldown:    %6d\n", stats.OnCooldown)
	fmt.Println()

	// Average rating
	if stats.AverageRating > 0 {
		fmt.Printf("Average Rating: %.1f/10.0\n", stats.AverageRating)
		fmt.Println()
	}

	// Storage
	if stats.TotalSize > 0 {
		sizeGB := float64(stats.TotalSize) / (1024 * 1024 * 1024)
		fmt.Printf("Total Storage: %.2f GB\n", sizeGB)
		fmt.Println()
	}

	// Detailed stats
	if detailed && len(stats.TopGenres) > 0 {
		fmt.Println("Top Genres")
		fmt.Println("──────────")

		// Sort genres by count
		type genreCount struct {
			genre string
			count int
		}
		var genres []genreCount
		for genre, count := range stats.TopGenres {
			genres = append(genres, genreCount{genre, count})
		}

		// Simple bubble sort
		for i := 0; i < len(genres); i++ {
			for j := i + 1; j < len(genres); j++ {
				if genres[i].count < genres[j].count {
					genres[i], genres[j] = genres[j], genres[i]
				}
			}
		}

		// Show top 10
		maxGenres := 10
		if len(genres) < maxGenres {
			maxGenres = len(genres)
		}
		for i := 0; i < maxGenres; i++ {
			fmt.Printf("  %-20s %4d\n", genres[i].genre, genres[i].count)
		}
		fmt.Println()
	}
}
