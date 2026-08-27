package filmroll

import (
	"context"
	"errors"
	"fmt"
	"time"

	db "github.com/C-A1R/exponia/backend/internal/database/generated"
	"github.com/jackc/pgx/v5"
	"github.com/jackc/pgx/v5/pgtype"
	"github.com/jackc/pgx/v5/pgxpool"
)

var ErrNotFound = errors.New("film roll not found")

type FilmRollStatus string

const (
	FilmRollStatusUnused    FilmRollStatus = "unused"
	FilmRollStatusReady     FilmRollStatus = "ready"
	FilmRollStatusInUse     FilmRollStatus = "in_use"
	FilmRollStatusExposed   FilmRollStatus = "exposed"
	FilmRollStatusDeveloped FilmRollStatus = "developed"
)

type FilmRollFilmStock struct {
	ID           int64  `json:"id"`
	Manufacturer string `json:"manufacturer"`
	Name         string `json:"name"`
	ISO          int32  `json:"iso"`
}

type FilmRollFormat struct {
	ID   int64  `json:"id"`
	Code string `json:"code"`
	Name string `json:"name"`
}

type FilmRollCamera struct {
	ID           int64  `json:"id"`
	Manufacturer string `json:"manufacturer"`
	Model        string `json:"model"`
}

type FilmRoll struct {
	ID          int64             `json:"id"`
	FilmStock   FilmRollFilmStock `json:"film_stock"`
	Format      FilmRollFormat    `json:"format"`
	Camera      *FilmRollCamera   `json:"camera"`
	ExposureISO int32             `json:"exposure_iso"`
	Status      FilmRollStatus    `json:"status"`
	CreatedAt   time.Time         `json:"created_at"`
}

type Repository struct {
	queries *db.Queries
}

func NewRepository(pool *pgxpool.Pool) *Repository {
	return &Repository{
		queries: db.New(pool),
	}
}

func fromGetRow(value db.GetFilmRollByIDRow) FilmRoll {
	var camera *FilmRollCamera

	if value.CameraID.Valid {
		camera = &FilmRollCamera{
			ID:           value.CameraID.Int64,
			Manufacturer: value.CameraManufacturer.String,
			Model:        value.CameraModel.String,
		}
	}

	return FilmRoll{
		ID: value.ID,

		FilmStock: FilmRollFilmStock{
			ID:           value.FilmStockID,
			Manufacturer: value.FilmStockManufacturer,
			Name:         value.FilmStockName,
			ISO:          value.FilmStockIso,
		},

		Format: FilmRollFormat{
			ID:   value.FormatID,
			Code: value.FormatCode,
			Name: value.FormatName,
		},

		Camera:      camera,
		ExposureISO: value.ExposureIso,
		Status:      FilmRollStatus(value.Status),
		CreatedAt:   value.CreatedAt.Time,
	}
}

func fromListRow(value db.ListFilmRollsRow) FilmRoll {
	var camera *FilmRollCamera

	if value.CameraID.Valid {
		camera = &FilmRollCamera{
			ID:           value.CameraID.Int64,
			Manufacturer: value.CameraManufacturer.String,
			Model:        value.CameraModel.String,
		}
	}

	return FilmRoll{
		ID: value.ID,

		FilmStock: FilmRollFilmStock{
			ID:           value.FilmStockID,
			Manufacturer: value.FilmStockManufacturer,
			Name:         value.FilmStockName,
			ISO:          value.FilmStockIso,
		},

		Format: FilmRollFormat{
			ID:   value.FormatID,
			Code: value.FormatCode,
			Name: value.FormatName,
		},

		Camera:      camera,
		ExposureISO: value.ExposureIso,
		Status:      FilmRollStatus(value.Status),
		CreatedAt:   value.CreatedAt.Time,
	}
}

func (r *Repository) Create(
	ctx context.Context,
	userID int64,
	filmStockID int64,
	formatID int64,
) (FilmRoll, error) {
	id, err := r.queries.CreateFilmRoll(
		ctx,
		db.CreateFilmRollParams{
			UserID:      userID,
			FilmStockID: filmStockID,
			FormatID:    formatID,
		},
	)
	if err != nil {
		return FilmRoll{}, fmt.Errorf("create film roll: %w", err)
	}

	return r.GetByID(ctx, userID, id)
}

func (r *Repository) GetByID(
	ctx context.Context,
	userID int64,
	filmRollID int64,
) (FilmRoll, error) {
	value, err := r.queries.GetFilmRollByID(
		ctx,
		db.GetFilmRollByIDParams{
			ID:     filmRollID,
			UserID: userID,
		},
	)
	if err != nil {
		if errors.Is(err, pgx.ErrNoRows) {
			return FilmRoll{}, ErrNotFound
		}

		return FilmRoll{}, fmt.Errorf(
			"get film roll by id: %w",
			err,
		)
	}

	return fromGetRow(value), nil
}

func (r *Repository) List(
	ctx context.Context,
	userID int64,
) ([]FilmRoll, error) {
	values, err := r.queries.ListFilmRolls(ctx, userID)
	if err != nil {
		return nil, fmt.Errorf("list film rolls: %w", err)
	}

	rolls := make([]FilmRoll, 0, len(values))

	for _, value := range values {
		rolls = append(rolls, fromListRow(value))
	}

	return rolls, nil
}

func (r *Repository) UpdateStatus(
	ctx context.Context,
	userID int64,
	filmRollID int64,
	status FilmRollStatus,
) (FilmRoll, error) {
	updatedID, err := r.queries.UpdateFilmRollStatus(
		ctx,
		db.UpdateFilmRollStatusParams{
			ID:     filmRollID,
			UserID: userID,
			Status: string(status),
		},
	)
	if err != nil {
		if errors.Is(err, pgx.ErrNoRows) {
			return FilmRoll{}, ErrNotFound
		}

		return FilmRoll{}, fmt.Errorf(
			"update film roll status: %w",
			err,
		)
	}

	return r.GetByID(ctx, userID, updatedID)
}

func (r *Repository) UpdateExposureISO(
	ctx context.Context,
	userID int64,
	filmRollID int64,
	exposureISO int32,
) (FilmRoll, error) {
	updatedID, err := r.queries.UpdateFilmRollExposureISO(
		ctx,
		db.UpdateFilmRollExposureISOParams{
			ID:          filmRollID,
			UserID:      userID,
			ExposureIso: exposureISO,
		},
	)
	if err != nil {
		if errors.Is(err, pgx.ErrNoRows) {
			return FilmRoll{}, ErrNotFound
		}

		return FilmRoll{}, fmt.Errorf(
			"update film roll exposure iso: %w",
			err,
		)
	}

	return r.GetByID(ctx, userID, updatedID)
}

func (r *Repository) UpdateCamera(
	ctx context.Context,
	userID int64,
	filmRollID int64,
	cameraID *int64,
) (FilmRoll, error) {
	var dbCameraID pgtype.Int8

	if cameraID != nil {
		dbCameraID = pgtype.Int8{
			Int64: *cameraID,
			Valid: true,
		}
	}

	updatedID, err := r.queries.UpdateFilmRollCamera(
		ctx,
		db.UpdateFilmRollCameraParams{
			ID:       filmRollID,
			UserID:   userID,
			CameraID: dbCameraID,
		},
	)
	if err != nil {
		if errors.Is(err, pgx.ErrNoRows) {
			return FilmRoll{}, ErrNotFound
		}

		return FilmRoll{}, fmt.Errorf(
			"update film roll camera: %w",
			err,
		)
	}

	return r.GetByID(ctx, userID, updatedID)
}

func (r *Repository) Delete(
	ctx context.Context,
	userID int64,
	filmRollID int64,
) error {
	rowsAffected, err := r.queries.DeleteFilmRoll(
		ctx,
		db.DeleteFilmRollParams{
			ID:     filmRollID,
			UserID: userID,
		},
	)
	if err != nil {
		return fmt.Errorf("delete film roll: %w", err)
	}

	if rowsAffected == 0 {
		return ErrNotFound
	}

	return nil
}
