package camera

import (
	"context"
	"errors"
	"fmt"
	"time"

	"github.com/jackc/pgx/v5"
	"github.com/jackc/pgx/v5/pgxpool"
)

var ErrNotFound = errors.New("camera not found")

type Camera struct {
	ID           int64     `json:"id"`
	Manufacturer string    `json:"manufacturer"`
	Model        string    `json:"model"`
	CreatedAt    time.Time `json:"created_at"`
}

type Repository struct {
	db *pgxpool.Pool
}

func NewRepository(db *pgxpool.Pool) *Repository {
	return &Repository{db: db}
}

func (r *Repository) CreateCamera(
	ctx context.Context,
	manufacturer string,
	model string,
) (Camera, error) {
	const query = `
		INSERT INTO cameras (manufacturer, model)
		VALUES ($1, $2)
		RETURNING id, manufacturer, model, created_at
	`

	var camera Camera

	err := r.db.QueryRow(
		ctx,
		query,
		manufacturer,
		model,
	).Scan(
		&camera.ID,
		&camera.Manufacturer,
		&camera.Model,
		&camera.CreatedAt,
	)

	if err != nil {
		return Camera{}, fmt.Errorf("insert camera: %w", err)
	}

	return camera, nil
}

func (r *Repository) ListCameras(ctx context.Context) ([]Camera, error) {
	const query = `
		SELECT *
		FROM cameras
		ORDER BY created_at DESC
	`

	rows, err := r.db.Query(ctx, query)
	if err != nil {
		return nil, fmt.Errorf("select cameras: %w", err)
	}
	defer rows.Close()

	cameras := make([]Camera, 0)

	for rows.Next() {
		var camera Camera

		if err := rows.Scan(
			&camera.ID,
			&camera.Manufacturer,
			&camera.Model,
			&camera.CreatedAt,
		); err != nil {
			return nil, fmt.Errorf("scan camera: %w", err)
		}

		cameras = append(cameras, camera)
	}

	if err := rows.Err(); err != nil {
		return nil, fmt.Errorf("iterate cameras: %w", err)
	}

	return cameras, nil
}

func (r *Repository) GetCameraById(ctx context.Context, id int64) (Camera, error) {
	const query = `
		SELECT *
		FROM cameras
		WHERE id = $1
	`

	var camera Camera

	err := r.db.QueryRow(ctx, query, id).Scan(
		&camera.ID,
		&camera.Manufacturer,
		&camera.Model,
		&camera.CreatedAt,
	)

	if err != nil {
		if errors.Is(err, pgx.ErrNoRows) {
			return Camera{}, ErrNotFound
		}
		return Camera{}, fmt.Errorf("select camera by id: %w", err)
	}

	return camera, nil
}

func (r *Repository) UpdateCamera(ctx context.Context, id int64, manufacturer string, model string) (Camera, error) {
	const query = `
		UPDATE cameras
		SET manufacturer = $2,
		    model = $3
		WHERE id = $1
		RETURNING id, manufacturer, model, created_at
	`

	var camera Camera

	err := r.db.QueryRow(ctx, query, id, manufacturer, model).Scan(
		&camera.ID,
		&camera.Manufacturer,
		&camera.Model,
		&camera.CreatedAt,
	)

	if err != nil {
		if errors.Is(err, pgx.ErrNoRows) {
			return Camera{}, ErrNotFound
		}
		return Camera{}, fmt.Errorf("update camera: %w", err)
	}

	return camera, nil
}

func (r *Repository) DeleteCamera(ctx context.Context, id int64) error {
	const query = `
		DELETE FROM cameras
		WHERE id = $1
	`

	cmdTag, err := r.db.Exec(ctx, query, id)
	if err != nil {
		return fmt.Errorf("delete camera: %w", err)
	}

	if cmdTag.RowsAffected() == 0 {
		return ErrNotFound
	}

	return nil
}
