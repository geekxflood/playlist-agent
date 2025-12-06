package repository

import (
	"context"
	"fmt"
	"time"

	"github.com/geekxflood/program-director/internal/database"
	"github.com/geekxflood/program-director/pkg/models"
)

// HistoryRepository handles play history persistence
type HistoryRepository struct {
	db database.DB
}

// NewHistoryRepository creates a new HistoryRepository
func NewHistoryRepository(db database.DB) *HistoryRepository {
	return &HistoryRepository{db: db}
}

// Create inserts a new play history record
func (r *HistoryRepository) Create(ctx context.Context, h *models.PlayHistory) error {
	if h.PlayedAt.IsZero() {
		h.PlayedAt = time.Now()
	}

	query := `
		INSERT INTO play_history (
			media_id, channel_id, theme_name, played_at, media_title, media_type
		) VALUES ($1, $2, $3, $4, $5, $6)
		RETURNING id
	`

	err := r.db.QueryRow(ctx, query,
		h.MediaID, h.ChannelID, h.ThemeName, h.PlayedAt, h.MediaTitle, h.MediaType,
	).Scan(&h.ID)

	return err
}

// List retrieves play history with optional filters
func (r *HistoryRepository) List(ctx context.Context, opts ListHistoryOptions) ([]models.PlayHistory, error) {
	query := `
		SELECT id, media_id, channel_id, theme_name, played_at, media_title, media_type
		FROM play_history WHERE 1=1
	`
	args := make([]interface{}, 0)
	argIndex := 1

	if opts.MediaID > 0 {
		query += fmt.Sprintf(" AND media_id = $%d", argIndex)
		args = append(args, opts.MediaID)
		argIndex++
	}

	if opts.ChannelID != "" {
		query += fmt.Sprintf(" AND channel_id = $%d", argIndex)
		args = append(args, opts.ChannelID)
		argIndex++
	}

	if opts.ThemeName != "" {
		query += fmt.Sprintf(" AND theme_name = $%d", argIndex)
		args = append(args, opts.ThemeName)
		argIndex++
	}

	if !opts.Since.IsZero() {
		query += fmt.Sprintf(" AND played_at >= $%d", argIndex)
		args = append(args, opts.Since)
		argIndex++
	}

	if !opts.Until.IsZero() {
		query += fmt.Sprintf(" AND played_at <= $%d", argIndex)
		args = append(args, opts.Until)
		argIndex++
	}

	// Order by played_at descending by default
	query += " ORDER BY played_at DESC"

	if opts.Limit > 0 {
		query += fmt.Sprintf(" LIMIT $%d", argIndex)
		args = append(args, opts.Limit)
		argIndex++
	}

	if opts.Offset > 0 {
		query += fmt.Sprintf(" OFFSET $%d", argIndex)
		args = append(args, opts.Offset)
	}

	rows, err := r.db.Query(ctx, query, args...)
	if err != nil {
		return nil, err
	}
	defer func() { _ = rows.Close() }()

	var history []models.PlayHistory
	for rows.Next() {
		var h models.PlayHistory
		err := rows.Scan(
			&h.ID, &h.MediaID, &h.ChannelID, &h.ThemeName, &h.PlayedAt, &h.MediaTitle, &h.MediaType,
		)
		if err != nil {
			return nil, err
		}
		history = append(history, h)
	}

	return history, rows.Err()
}

// Count returns the total number of play history records
func (r *HistoryRepository) Count(ctx context.Context, opts ListHistoryOptions) (int64, error) {
	query := "SELECT COUNT(*) FROM play_history WHERE 1=1"
	args := make([]interface{}, 0)
	argIndex := 1

	if opts.MediaID > 0 {
		query += fmt.Sprintf(" AND media_id = $%d", argIndex)
		args = append(args, opts.MediaID)
		argIndex++
	}

	if opts.ChannelID != "" {
		query += fmt.Sprintf(" AND channel_id = $%d", argIndex)
		args = append(args, opts.ChannelID)
		argIndex++
	}

	if opts.ThemeName != "" {
		query += fmt.Sprintf(" AND theme_name = $%d", argIndex)
		args = append(args, opts.ThemeName)
		argIndex++
	}

	if !opts.Since.IsZero() {
		query += fmt.Sprintf(" AND played_at >= $%d", argIndex)
		args = append(args, opts.Since)
		// argIndex++ not needed as it's the last use
	}

	var count int64
	err := r.db.QueryRow(ctx, query, args...).Scan(&count)
	return count, err
}

// ListHistoryOptions provides filtering options for List
type ListHistoryOptions struct {
	MediaID   int64
	ChannelID string
	ThemeName string
	Since     time.Time
	Until     time.Time
	Limit     int
	Offset    int
}
