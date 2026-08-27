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

type Repository struct {
	queries *db.Queries
}

func NewRepository(pool *pgxpool.Pool) *Repository {
	return &Repository{queries: db.New(pool)}
}

func fromDBCamera(value db.Camera) Camera {
	return Camera{
		ID:           value.ID,
		Manufacturer: value.Manufacturer,
		Model:        value.Model,
		CreatedAt:    value.CreatedAt.Time,
	}
}

func (r *Repository) CreateCamera(
	ctx context.Context,
	userID int64,
	manufacturer string,
	model string,
) (Camera, error) {
	camera, err := r.queries.CreateCamera(
		ctx,
		db.CreateCameraParams{
			UserID:       userID,
			Manufacturer: manufacturer,
			Model:        model,
		},
	)
	if err != nil {
		return Camera{}, fmt.Errorf("create camera: %w", err)
	}

	return fromDBCamera(camera), nil
}

func (r *Repository) ListCameras(
	ctx context.Context,
	userID int64,
) ([]Camera, error) {
	values, err := r.queries.ListCameras(ctx, userID)
	if err != nil {
		return nil, fmt.Errorf("list cameras: %w", err)
	}

	cameras := make([]Camera, 0, len(values))
	for _, value := range values {
		cameras = append(cameras, fromDBCamera(value))
	}

	return cameras, nil
}

func (r *Repository) GetCameraByID(
	ctx context.Context,
	userID int64,
	cameraID int64,
) (Camera, error) {
	camera, err := r.queries.GetCameraByID(
		ctx,
		db.GetCameraByIDParams{
			ID:     cameraID,
			UserID: userID,
		},
	)
	if err != nil {
		if errors.Is(err, pgx.ErrNoRows) {
			return Camera{}, ErrNotFound
		}

		return Camera{}, fmt.Errorf("get camera by id: %w", err)
	}

	return fromDBCamera(camera), nil
}

func (r *Repository) UpdateCamera(
	ctx context.Context,
	userID int64,
	cameraID int64,
	manufacturer string,
	model string,
) (Camera, error) {
	camera, err := r.queries.UpdateCamera(
		ctx,
		db.UpdateCameraParams{
			ID:           cameraID,
			UserID:       userID,
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

	return fromDBCamera(camera), nil
}

func (r *Repository) DeleteCamera(
	ctx context.Context,
	userID int64,
	cameraID int64,
) error {
	rowsAffected, err := r.queries.DeleteCamera(
		ctx,
		db.DeleteCameraParams{
			ID:     cameraID,
			UserID: userID,
		},
	)
	if err != nil {
		return fmt.Errorf("delete camera: %w", err)
	}

	if rowsAffected == 0 {
		return ErrNotFound
	}

	return nil
}
