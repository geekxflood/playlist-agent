package cmd

import (
	"context"
	"os"
	"os/signal"
	"syscall"

	"github.com/spf13/cobra"
)

var (
	syncMovies  bool
	syncSeries  bool
	syncCleanup bool
)

// syncCmd represents the sync command
var syncCmd = &cobra.Command{
	Use:   "sync",
	Short: "Sync media catalog from Radarr/Sonarr",
	Long: `Synchronize the local media catalog with Radarr and Sonarr.

This command fetches all media metadata from your media management
applications and stores it in the local database for fast querying
during playlist generation.

Examples:
  # Sync all media (movies and series)
  program-director sync

  # Sync only movies
  program-director sync --movies

  # Sync only series (TV shows and anime)
  program-director sync --series

  # Sync and cleanup removed media
  program-director sync --cleanup`,
	RunE: runSync,
}

func init() {
	syncCmd.Flags().BoolVar(&syncMovies, "movies", false, "sync only movies from Radarr")
	syncCmd.Flags().BoolVar(&syncSeries, "series", false, "sync only series from Sonarr")
	syncCmd.Flags().BoolVar(&syncCleanup, "cleanup", false, "remove media no longer in source")
}

func runSync(cmd *cobra.Command, args []string) error {
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

	// Default to syncing everything if no specific flags
	syncAll := !syncMovies && !syncSeries
	if syncAll {
		syncMovies = true
		syncSeries = true
	}

	logger.Info("starting media sync",
		"movies", syncMovies,
		"series", syncSeries,
		"cleanup", syncCleanup,
		"radarr_url", cfg.Radarr.URL,
		"sonarr_url", cfg.Sonarr.URL,
	)

	logger.Debug("initializing sync services")
	// TODO: Initialize database and sync service
	// This will be implemented when media sync service is wired up
	logger.Warn("sync service not yet implemented - placeholder only")

	if syncMovies {
		logger.Info("syncing movies from Radarr",
			"url", cfg.Radarr.URL,
		)
		logger.Debug("fetching movie list from Radarr API")
		select {
		case <-ctx.Done():
			return ctx.Err()
		default:
			// syncService.SyncMovies(ctx)
			logger.Debug("movie sync would happen here")
		}
		logger.Info("movie sync completed", "count", 0)
	}

	if syncSeries {
		logger.Info("syncing series from Sonarr",
			"url", cfg.Sonarr.URL,
		)
		logger.Debug("fetching series list from Sonarr API")
		select {
		case <-ctx.Done():
			return ctx.Err()
		default:
			// syncService.SyncSeries(ctx)
			logger.Debug("series sync would happen here")
		}
		logger.Info("series sync completed", "count", 0)
	}

	if syncCleanup {
		logger.Info("cleaning up removed media")
		logger.Debug("checking for media no longer in source")
		// syncService.Cleanup(ctx)
		logger.Debug("cleanup would happen here")
		logger.Info("cleanup completed", "removed", 0)
	}

	logger.Info("media sync complete",
		"movies_synced", 0,
		"series_synced", 0,
		"items_cleaned", 0,
	)
	return nil
}
