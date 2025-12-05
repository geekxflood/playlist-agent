package scheduler

import (
	"context"
	"log/slog"
	"os"
	"testing"
	"time"

	"github.com/geekxflood/program-director/internal/config"
)

// mockGenerator is a mock implementation of the playlist generator
type mockGenerator struct {
	generateAllCalled bool
	themes            []config.ThemeConfig
	dryRun            bool
}

func (m *mockGenerator) GenerateAll(ctx context.Context, themes []config.ThemeConfig, dryRun bool) ([]interface{}, error) {
	m.generateAllCalled = true
	m.themes = themes
	m.dryRun = dryRun
	return nil, nil
}

func TestNewScheduler(t *testing.T) {
	cfg := &Config{
		Schedule: "0 2 * * *",
		DryRun:   false,
	}

	themes := []config.ThemeConfig{
		{Name: "test-theme", ChannelID: "ch1"},
	}

	logger := slog.New(slog.NewTextHandler(os.Stdout, nil))

	sched, err := NewScheduler(cfg, nil, themes, logger)
	if err != nil {
		t.Fatalf("expected no error, got %v", err)
	}

	if sched == nil {
		t.Fatal("expected non-nil scheduler")
	}

	if len(sched.themes) != 1 {
		t.Errorf("expected 1 theme, got %d", len(sched.themes))
	}

	if sched.logger == nil {
		t.Error("expected non-nil logger")
	}

	if sched.cron == nil {
		t.Error("expected non-nil cron")
	}
}

func TestNewSchedulerDefaultSchedule(t *testing.T) {
	cfg := &Config{} // Empty schedule

	themes := []config.ThemeConfig{}
	logger := slog.New(slog.NewTextHandler(os.Stdout, nil))

	sched, err := NewScheduler(cfg, nil, themes, logger)
	if err != nil {
		t.Fatalf("expected no error, got %v", err)
	}

	if sched == nil {
		t.Fatal("expected non-nil scheduler")
	}
}

func TestGetStatus(t *testing.T) {
	cfg := &Config{
		Schedule: "0 2 * * *",
	}

	themes := []config.ThemeConfig{
		{Name: "test-theme", ChannelID: "ch1"},
	}

	logger := slog.New(slog.NewTextHandler(os.Stdout, nil))

	sched, err := NewScheduler(cfg, nil, themes, logger)
	if err != nil {
		t.Fatalf("expected no error, got %v", err)
	}

	status := sched.GetStatus()

	if status == nil {
		t.Fatal("expected non-nil status")
	}

	if themes, ok := status["themes"].(int); !ok || themes != 1 {
		t.Errorf("expected themes count 1, got %v", status["themes"])
	}

	if jobs, ok := status["jobs"].(int); !ok || jobs != 0 {
		t.Errorf("expected jobs count 0 (not started), got %v", status["jobs"])
	}
}

func TestStop(t *testing.T) {
	cfg := &Config{
		Schedule: "0 2 * * *",
	}

	themes := []config.ThemeConfig{}
	logger := slog.New(slog.NewTextHandler(os.Stdout, nil))

	sched, err := NewScheduler(cfg, nil, themes, logger)
	if err != nil {
		t.Fatalf("expected no error, got %v", err)
	}

	// Stop should work even if not started
	err = sched.Stop()
	if err != nil {
		t.Errorf("expected no error stopping scheduler, got %v", err)
	}
}

func TestGetNextRun(t *testing.T) {
	cfg := &Config{
		Schedule: "0 2 * * *",
	}

	themes := []config.ThemeConfig{}
	logger := slog.New(slog.NewTextHandler(os.Stdout, nil))

	sched, err := NewScheduler(cfg, nil, themes, logger)
	if err != nil {
		t.Fatalf("expected no error, got %v", err)
	}

	// Before starting, next run should be zero time
	nextRun := sched.GetNextRun()
	if !nextRun.IsZero() {
		t.Error("expected zero time before starting scheduler")
	}
}

func TestSchedulerStartStop(t *testing.T) {
	// Use a schedule that runs every second for testing
	cfg := &Config{
		Schedule: "* * * * * *", // Every second
	}

	themes := []config.ThemeConfig{
		{Name: "test-theme", ChannelID: "ch1"},
	}

	logger := slog.New(slog.NewTextHandler(os.Stdout, nil))

	sched, err := NewScheduler(cfg, nil, themes, logger)
	if err != nil {
		t.Fatalf("expected no error, got %v", err)
	}

	ctx, cancel := context.WithCancel(context.Background())
	defer cancel()

	// Start scheduler in goroutine
	go func() {
		sched.Start(ctx, cfg.Schedule, false)
	}()

	// Give it a moment to start
	time.Sleep(100 * time.Millisecond)

	// Cancel context to stop
	cancel()

	// Give it a moment to stop
	time.Sleep(100 * time.Millisecond)

	// Test should complete without hanging
}
