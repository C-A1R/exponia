package lens

import (
	"context"
	"errors"
	"fmt"

	db "github.com/C-A1R/exponia/backend/internal/database/generated"
	"github.com/jackc/pgx/v5"
	"github.com/jackc/pgx/v5/pgxpool"
)

var ErrNotFound = errors.New("lens not found")

type Repository struct {
	queries *db.Queries
}

func NewRepository(pool *pgxpool.Pool) *Repository {
	return &Repository{
		queries: db.New(pool),
	}
}

func fromDBLens(value db.Lense) Lens {
	return Lens{
		ID:            value.ID,
		Manufacturer:  value.Manufacturer,
		Model:         value.Model,
		FocalLengthMm: value.FocalLengthMm,
		MaxAperture:   value.MaxAperture,
		CreatedAt:     value.CreatedAt.Time,
	}
}

func (r *Repository) Create(
	ctx context.Context,
	manufacturer string,
	model string,
	focalLengthMm int32,
	maxAperture float64,
) (Lens, error) {
	value, err := r.queries.CreateLens(
		ctx,
		db.CreateLensParams{
			Manufacturer:  manufacturer,
			Model:         model,
			FocalLengthMm: focalLengthMm,
			MaxAperture:   maxAperture,
		},
	)
	if err != nil {
		return Lens{}, fmt.Errorf("create lens: %w", err)
	}

	return fromDBLens(value), nil
}

func (r *Repository) List(ctx context.Context) ([]Lens, error) {
	values, err := r.queries.ListLenses(ctx)
	if err != nil {
		return nil, fmt.Errorf("list lenses: %w", err)
	}

	lenses := make([]Lens, 0, len(values))

	for _, value := range values {
		lenses = append(lenses, fromDBLens(value))
	}

	return lenses, nil
}

func (r *Repository) GetByID(
	ctx context.Context,
	id int64,
) (Lens, error) {
	value, err := r.queries.GetLensByID(ctx, id)
	if err != nil {
		if errors.Is(err, pgx.ErrNoRows) {
			return Lens{}, ErrNotFound
		}

		return Lens{}, fmt.Errorf("get lens by id: %w", err)
	}

	return fromDBLens(value), nil
}

func (r *Repository) Update(
	ctx context.Context,
	id int64,
	manufacturer string,
	model string,
	focalLengthMm int32,
	maxAperture float64,
) (Lens, error) {
	value, err := r.queries.UpdateLens(
		ctx,
		db.UpdateLensParams{
			ID:            id,
			Manufacturer:  manufacturer,
			Model:         model,
			FocalLengthMm: focalLengthMm,
			MaxAperture:   maxAperture,
		},
	)
	if err != nil {
		if errors.Is(err, pgx.ErrNoRows) {
			return Lens{}, ErrNotFound
		}

		return Lens{}, fmt.Errorf("update lens: %w", err)
	}

	return fromDBLens(value), nil
}

func (r *Repository) Delete(
	ctx context.Context,
	id int64,
) error {
	rowsAffected, err := r.queries.DeleteLens(ctx, id)
	if err != nil {
		return fmt.Errorf("delete lens: %w", err)
	}

	if rowsAffected == 0 {
		return ErrNotFound
	}

	return nil
}
