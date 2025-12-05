package scheduler

import (
	"context"
	"fmt"
	"log/slog"
	"time"

	"github.com/robfig/cron/v3"

	"github.com/geekxflood/program-director/internal/config"
	"github.com/geekxflood/program-director/internal/services/playlist"
)

// Scheduler handles automated playlist generation on a cron schedule
type Scheduler struct {
	cron      *cron.Cron
	generator *playlist.Generator
	themes    []config.ThemeConfig
	logger    *slog.Logger
}

// Config holds scheduler configuration
type Config struct {
	// Schedule defines when to run generation (cron format)
	// Default: "0 2 * * *" (daily at 2 AM)
	Schedule string
	// DryRun enables dry-run mode (no actual changes)
	DryRun bool
}

// NewScheduler creates a new scheduler instance
func NewScheduler(
	cfg *Config,
	generator *playlist.Generator,
	themes []config.ThemeConfig,
	logger *slog.Logger,
) (*Scheduler, error) {
	if cfg.Schedule == "" {
		cfg.Schedule = "0 2 * * *" // Default: daily at 2 AM
	}

	// Create cron with second precision and logging
	cronLogger := cron.VerbosePrintfLogger(
		slog.NewLogLogger(logger.Handler(), slog.LevelInfo),
	)

	c := cron.New(
		cron.WithLogger(cronLogger),
		cron.WithChain(
			cron.Recover(cronLogger),
		),
	)

	return &Scheduler{
		cron:      c,
		generator: generator,
		themes:    themes,
		logger:    logger,
	}, nil
}

// Start starts the scheduler
func (s *Scheduler) Start(ctx context.Context, schedule string, dryRun bool) error {
	s.logger.Info("starting scheduler",
		"schedule", schedule,
		"themes", len(s.themes),
		"dry_run", dryRun,
	)

	// Add generation job
	_, err := s.cron.AddFunc(schedule, func() {
		s.runGeneration(context.Background(), dryRun)
	})
	if err != nil {
		return fmt.Errorf("failed to add cron job: %w", err)
	}

	// Start cron scheduler
	s.cron.Start()

	s.logger.Info("scheduler started successfully")

	// Block until context cancelled
	<-ctx.Done()

	s.logger.Info("stopping scheduler")
	return s.Stop()
}

// Stop stops the scheduler
func (s *Scheduler) Stop() error {
	ctx := s.cron.Stop()
	<-ctx.Done()
	s.logger.Info("scheduler stopped")
	return nil
}

// RunNow triggers an immediate generation run
func (s *Scheduler) RunNow(ctx context.Context, dryRun bool) error {
	s.logger.Info("manual generation triggered")
	s.runGeneration(ctx, dryRun)
	return nil
}

// runGeneration executes playlist generation for all themes
func (s *Scheduler) runGeneration(ctx context.Context, dryRun bool) {
	start := time.Now()

	s.logger.Info("scheduled generation started",
		"themes", len(s.themes),
		"dry_run", dryRun,
	)

	results, err := s.generator.GenerateAll(ctx, s.themes, dryRun)
	if err != nil {
		s.logger.Error("generation failed", "error", err)
		return
	}

	// Log results
	var successCount, failCount int
	for _, result := range results {
		if result.Error != nil {
			failCount++
			s.logger.Error("theme generation failed",
				"theme", result.ThemeName,
				"error", result.Error,
			)
		} else if result.Generated {
			successCount++
			s.logger.Info("theme generation succeeded",
				"theme", result.ThemeName,
				"items", result.ItemCount,
				"duration", result.Duration,
			)
		} else {
			s.logger.Warn("theme generation skipped",
				"theme", result.ThemeName,
			)
		}
	}

	s.logger.Info("scheduled generation complete",
		"total", len(results),
		"success", successCount,
		"failed", failCount,
		"duration", time.Since(start),
	)
}

// GetNextRun returns the next scheduled run time
func (s *Scheduler) GetNextRun() time.Time {
	entries := s.cron.Entries()
	if len(entries) == 0 {
		return time.Time{}
	}
	return entries[0].Next
}

// GetStatus returns scheduler status information
func (s *Scheduler) GetStatus() map[string]interface{} {
	entries := s.cron.Entries()
	var nextRun time.Time
	if len(entries) > 0 {
		nextRun = entries[0].Next
	}

	return map[string]interface{}{
		"running":    len(entries) > 0,
		"themes":     len(s.themes),
		"jobs":       len(entries),
		"next_run":   nextRun.Format(time.RFC3339),
		"next_run_in": time.Until(nextRun).String(),
	}
}
