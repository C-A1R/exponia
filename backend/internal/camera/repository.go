package camera

import (
	"context"
	"errors"
	"fmt"

	db "github.com/C-A1R/exponia/backend/internal/database/generated"
	"github.com/jackc/pgx/v5"
	"github.com/jackc/pgx/v5/pgxpool"
)

var ErrNotFound = errors.New("camera not found")

type Camera = db.Camera

type Repository struct {
	queries *db.Queries
}

func NewRepository(pool *pgxpool.Pool) *Repository {
	return &Repository{queries: db.New(pool)}
}

func (r *Repository) CreateCamera(
	ctx context.Context,
	manufacturer string,
	model string,
) (Camera, error) {
	camera, err := r.queries.CreateCamera(
		ctx,
		db.CreateCameraParams{
			Manufacturer: manufacturer,
			Model:        model,
		},
	)
	if err != nil {
		return Camera{}, fmt.Errorf("create camera: %w", err)
	}

	return camera, nil
}

func (r *Repository) ListCameras(ctx context.Context) ([]Camera, error) {
	cameras, err := r.queries.ListCameras(ctx)
	if err != nil {
		return nil, fmt.Errorf("list cameras: %w", err)
	}

	if cameras == nil {
		cameras = make([]Camera, 0)
	}

	return cameras, nil
}

func (r *Repository) GetCameraById(ctx context.Context, id int64) (Camera, error) {
	camera, err := r.queries.GetCameraByID(ctx, id)
	if err != nil {
		if errors.Is(err, pgx.ErrNoRows) {
			return Camera{}, ErrNotFound
		}

		return Camera{}, fmt.Errorf("get camera by id: %w", err)
	}

	return camera, nil
}

func (r *Repository) UpdateCamera(ctx context.Context, id int64, manufacturer string, model string) (Camera, error) {
	camera, err := r.queries.UpdateCamera(
		ctx,
		db.UpdateCameraParams{
			ID:           id,
			Manufacturer: manufacturer,
			Model:        model,
		},
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
	rowsAffected, err := r.queries.DeleteCamera(ctx, id)
	if err != nil {
		return fmt.Errorf("delete camera: %w", err)
	}

	if rowsAffected == 0 {
		return ErrNotFound
	}

	return nil
}
