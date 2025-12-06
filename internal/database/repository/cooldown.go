// Package repository provides database repository implementations for media, history, and cooldowns.
package repository

import (
	"context"
	"fmt"
	"time"

	"github.com/geekxflood/program-director/internal/database"
	"github.com/geekxflood/program-director/pkg/models"
)

// CooldownRepository handles media cooldown persistence
type CooldownRepository struct {
	db database.DB
}

// NewCooldownRepository creates a new CooldownRepository
func NewCooldownRepository(db database.DB) *CooldownRepository {
	return &CooldownRepository{db: db}
}

// Upsert creates or updates a cooldown record
func (r *CooldownRepository) Upsert(ctx context.Context, c *models.MediaCooldown) error {
	query := `
		INSERT INTO media_cooldowns (
			media_id, cooldown_days, last_played_at, can_replay_at, media_title, media_type
		) VALUES ($1, $2, $3, $4, $5, $6)
		ON CONFLICT (media_id) DO UPDATE SET
			cooldown_days = EXCLUDED.cooldown_days,
			last_played_at = EXCLUDED.last_played_at,
			can_replay_at = EXCLUDED.can_replay_at,
			media_title = EXCLUDED.media_title,
			media_type = EXCLUDED.media_type
		RETURNING id
	`

	err := r.db.QueryRow(ctx, query,
		c.MediaID, c.CooldownDays, c.LastPlayedAt, c.CanReplayAt, c.MediaTitle, c.MediaType,
	).Scan(&c.ID)

	return err
}

// List retrieves cooldowns with optional filters
func (r *CooldownRepository) List(ctx context.Context, opts ListCooldownOptions) ([]models.MediaCooldown, error) {
	query := `
		SELECT id, media_id, cooldown_days, last_played_at, can_replay_at, media_title, media_type
		FROM media_cooldowns WHERE 1=1
	`
	args := make([]interface{}, 0)
	argIndex := 1

	if opts.MediaType != "" {
		query += fmt.Sprintf(" AND media_type = $%d", argIndex)
		args = append(args, opts.MediaType)
		argIndex++
	}

	if opts.ActiveOnly {
		query += fmt.Sprintf(" AND can_replay_at > $%d", argIndex)
		args = append(args, time.Now())
		argIndex++
	}

	if opts.ExpiredOnly {
		query += fmt.Sprintf(" AND can_replay_at <= $%d", argIndex)
		args = append(args, time.Now())
		argIndex++
	}

	query += " ORDER BY can_replay_at"

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

	var cooldowns []models.MediaCooldown
	for rows.Next() {
		var c models.MediaCooldown
		err := rows.Scan(
			&c.ID, &c.MediaID, &c.CooldownDays, &c.LastPlayedAt, &c.CanReplayAt, &c.MediaTitle, &c.MediaType,
		)
		if err != nil {
			return nil, err
		}
		cooldowns = append(cooldowns, c)
	}

	return cooldowns, rows.Err()
}

// GetActiveCooldownMediaIDs returns IDs of media currently on cooldown
func (r *CooldownRepository) GetActiveCooldownMediaIDs(ctx context.Context) ([]int64, error) {
	rows, err := r.db.Query(ctx,
		"SELECT media_id FROM media_cooldowns WHERE can_replay_at > $1",
		time.Now(),
	)
	if err != nil {
		return nil, err
	}
	defer func() { _ = rows.Close() }()

	var ids []int64
	for rows.Next() {
		var id int64
		if err := rows.Scan(&id); err != nil {
			return nil, err
		}
		ids = append(ids, id)
	}
	return ids, rows.Err()
}

// CountActive returns the number of media currently on cooldown
func (r *CooldownRepository) CountActive(ctx context.Context) (int64, error) {
	var count int64
	err := r.db.QueryRow(ctx,
		"SELECT COUNT(*) FROM media_cooldowns WHERE can_replay_at > $1",
		time.Now(),
	).Scan(&count)
	return count, err
}

// ListCooldownOptions provides filtering options for List
type ListCooldownOptions struct {
	MediaType   models.MediaType
	ActiveOnly  bool
	ExpiredOnly bool
	Limit       int
	Offset      int
}
