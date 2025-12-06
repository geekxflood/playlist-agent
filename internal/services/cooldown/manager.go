// Package cooldown provides media cooldown management to prevent recent replays.
package cooldown

import (
	"context"
	"log/slog"
	"time"

	"github.com/geekxflood/program-director/internal/config"
	"github.com/geekxflood/program-director/internal/database/repository"
	"github.com/geekxflood/program-director/pkg/models"
)

// Manager handles media cooldown tracking
type Manager struct {
	cooldownRepo *repository.CooldownRepository
	historyRepo  *repository.HistoryRepository
	config       *config.CooldownConfig
	logger       *slog.Logger
}

// NewManager creates a new cooldown Manager
func NewManager(
	cooldownRepo *repository.CooldownRepository,
	historyRepo *repository.HistoryRepository,
	cfg *config.CooldownConfig,
	logger *slog.Logger,
) *Manager {
	return &Manager{
		cooldownRepo: cooldownRepo,
		historyRepo:  historyRepo,
		config:       cfg,
		logger:       logger,
	}
}

// RecordPlay records that a media item was played and sets its cooldown
func (m *Manager) RecordPlay(ctx context.Context, media *models.Media, channelID, themeName string) error {
	now := time.Now()

	// Create play history record
	history := &models.PlayHistory{
		MediaID:    media.ID,
		ChannelID:  channelID,
		ThemeName:  themeName,
		PlayedAt:   now,
		MediaTitle: media.Title,
		MediaType:  media.MediaType,
	}

	if err := m.historyRepo.Create(ctx, history); err != nil {
		return err
	}

	// Determine cooldown days based on media type
	cooldownDays := m.getCooldownDays(media.MediaType)

	// Create or update cooldown
	cooldown := &models.MediaCooldown{
		MediaID:      media.ID,
		CooldownDays: cooldownDays,
		LastPlayedAt: now,
		CanReplayAt:  now.AddDate(0, 0, cooldownDays),
		MediaTitle:   media.Title,
		MediaType:    media.MediaType,
	}

	if err := m.cooldownRepo.Upsert(ctx, cooldown); err != nil {
		return err
	}

	m.logger.Debug("recorded play and cooldown",
		"media_id", media.ID,
		"title", media.Title,
		"cooldown_days", cooldownDays,
		"can_replay_at", cooldown.CanReplayAt,
	)

	return nil
}

// GetActiveCooldownMediaIDs returns IDs of all media currently on cooldown
func (m *Manager) GetActiveCooldownMediaIDs(ctx context.Context) ([]int64, error) {
	return m.cooldownRepo.GetActiveCooldownMediaIDs(ctx)
}

// getCooldownDays returns the cooldown days for a media type
func (m *Manager) getCooldownDays(mediaType models.MediaType) int {
	switch mediaType {
	case models.MediaTypeMovie:
		return m.config.MovieDays
	case models.MediaTypeSeries:
		return m.config.SeriesDays
	case models.MediaTypeAnime:
		return m.config.AnimeDays
	default:
		return m.config.MovieDays
	}
}
