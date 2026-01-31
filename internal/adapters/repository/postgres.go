package repository

import (
	"context"
	"database/sql"

	"github.com/RiosHectorM/gopher-stream/internal/domain"

	_ "github.com/jackc/pgx/v5/stdlib" // Driver de alto rendimiento
)

type PostgresRepository struct {
	db *sql.DB
}

func NewPostgresRepository(connStr string) (*PostgresRepository, error) {
	db, err := sql.Open("pgx", connStr)
	if err != nil {
		return nil, err
	}
	return &PostgresRepository{db: db}, nil
}

func (r *PostgresRepository) UpdateLocation(ctx context.Context, event domain.Event) error {
	query := `
        INSERT INTO asset_events (asset_id, lat, long, payload, timestamp)
        VALUES ($1, $2, $3, $4, $5)
    `
	// Usamos Context para manejar timeouts (fundamental en sistemas críticos)
	_, err := r.db.ExecContext(ctx, query, event.AssetID, event.Lat, event.Long, event.Payload, event.Timestamp)
	return err
}

func (r *PostgresRepository) GetByID(ctx context.Context, id string) (domain.Asset, error) {
	var asset domain.Asset
	query := `SELECT id, type, status, last_update FROM assets WHERE id = $1`

	err := r.db.QueryRowContext(ctx, query, id).Scan(&asset.ID, &asset.Type, &asset.Status, &asset.LastUpdate)
	return asset, err
}

func (r *PostgresRepository) SaveToDLQ(ctx context.Context, event domain.Event, reason string) error {
    query := `INSERT INTO dead_letter_events (asset_id, payload, error_message) VALUES ($1, $2, $3)`
    _, err := r.db.ExecContext(ctx, query, event.AssetID, event.Payload, reason)
    return err
}