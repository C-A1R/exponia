package exposure

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

var (
	ErrNotFound          = errors.New("exposure not found")
	ErrFrameNotFound     = errors.New("frame not found")
	ErrCameraNotSelected = errors.New("camera not selected")
	ErrCameraNotFound    = errors.New("camera not found")
	ErrLensNotFound      = errors.New("lens not found")
	ErrCreateRejected    = errors.New("exposure creation rejected")
	ErrUpdateRejected    = errors.New("exposure update rejected")
)

type Repository struct {
	queries *db.Queries
}

func NewRepository(pool *pgxpool.Pool) *Repository {
	return &Repository{
		queries: db.New(pool),
	}
}

func toDBInt8(value *int64) pgtype.Int8 {
	if value == nil {
		return pgtype.Int8{}
	}

	return pgtype.Int8{
		Int64: *value,
		Valid: true,
	}
}

func toDBFloat8(value *float64) pgtype.Float8 {
	if value == nil {
		return pgtype.Float8{}
	}

	return pgtype.Float8{
		Float64: *value,
		Valid:   true,
	}
}

func toDBTimestamptz(value *time.Time) pgtype.Timestamptz {
	if value == nil {
		return pgtype.Timestamptz{}
	}

	return pgtype.Timestamptz{
		Time:  *value,
		Valid: true,
	}
}

func toDBText(value *string) pgtype.Text {
	if value == nil {
		return pgtype.Text{}
	}

	return pgtype.Text{
		String: *value,
		Valid:  true,
	}
}

func fromDBInt8(value pgtype.Int8) *int64 {
	if !value.Valid {
		return nil
	}

	result := value.Int64

	return &result
}

func fromDBFloat8(value pgtype.Float8) *float64 {
	if !value.Valid {
		return nil
	}

	result := value.Float64

	return &result
}

func fromDBTimestamptz(value pgtype.Timestamptz) *time.Time {
	if !value.Valid {
		return nil
	}

	result := value.Time

	return &result
}

func fromDBText(value pgtype.Text) *string {
	if !value.Valid {
		return nil
	}

	result := value.String

	return &result
}

func fromDBExposure(value db.Exposure) Exposure {
	return Exposure{
		ID:             value.ID,
		FrameID:        value.FrameID,
		ExposureIndex:  value.ExposureIndex,
		CameraID:       value.CameraID,
		LensID:         fromDBInt8(value.LensID),
		Aperture:       fromDBFloat8(value.Aperture),
		ShutterSpeedUS: fromDBInt8(value.ShutterSpeedUs),
		ShotAt:         fromDBTimestamptz(value.ShotAt),
		Note:           fromDBText(value.Note),
		CreatedAt:      value.CreatedAt.Time,
	}
}

func (r *Repository) Create(
	ctx context.Context,
	userID int64,
	frameID int64,
	input CreateInput,
) (Exposure, error) {
	dbLensID := toDBInt8(input.LensID)

	state, err := r.queries.GetExposureCreationState(
		ctx,
		db.GetExposureCreationStateParams{
			LensID:  dbLensID,
			UserID:  userID,
			FrameID: frameID,
		},
	)
	if err != nil {
		if errors.Is(err, pgx.ErrNoRows) {
			return Exposure{}, ErrFrameNotFound
		}

		return Exposure{}, fmt.Errorf("get exposure creation state: %w", err)
	}

	if !state.CameraID.Valid {
		return Exposure{}, ErrCameraNotSelected
	}

	if !state.LensAvailable.Valid || !state.LensAvailable.Bool {
		return Exposure{}, ErrLensNotFound
	}

	value, err := r.queries.CreateExposure(
		ctx,
		db.CreateExposureParams{
			LensID:         dbLensID,
			Aperture:       toDBFloat8(input.Aperture),
			ShutterSpeedUs: toDBInt8(input.ShutterSpeedUS),
			ShotAt:         toDBTimestamptz(input.ShotAt),
			Note:           toDBText(input.Note),
			FrameID:        frameID,
			UserID:         userID,
		},
	)
	if err != nil {
		if errors.Is(err, pgx.ErrNoRows) {
			return Exposure{}, ErrCreateRejected
		}

		return Exposure{}, fmt.Errorf("create exposure: %w", err)
	}

	return fromDBExposure(value), nil
}

func (r *Repository) GetByID(
	ctx context.Context,
	userID int64,
	exposureID int64,
) (Exposure, error) {
	value, err := r.queries.GetExposureByID(
		ctx,
		db.GetExposureByIDParams{
			ID:     exposureID,
			UserID: userID,
		},
	)
	if err != nil {
		if errors.Is(err, pgx.ErrNoRows) {
			return Exposure{}, ErrNotFound
		}

		return Exposure{}, fmt.Errorf("get exposure by id: %w", err)
	}

	return fromDBExposure(value), nil
}

func (r *Repository) List(
	ctx context.Context,
	userID int64,
	frameID int64,
) ([]Exposure, error) {
	_, err := r.queries.GetOwnedExposureFrameID(
		ctx,
		db.GetOwnedExposureFrameIDParams{
			FrameID: frameID,
			UserID:  userID,
		},
	)
	if err != nil {
		if errors.Is(err, pgx.ErrNoRows) {
			return nil, ErrFrameNotFound
		}

		return nil, fmt.Errorf("get frame for exposure list: %w", err)
	}

	values, err := r.queries.ListExposures(
		ctx,
		db.ListExposuresParams{
			FrameID: frameID,
			UserID:  userID,
		},
	)
	if err != nil {
		return nil, fmt.Errorf("list exposures: %w", err)
	}

	exposures := make([]Exposure, 0, len(values))
	for _, value := range values {
		exposures = append(exposures, fromDBExposure(value))
	}

	return exposures, nil
}

func (r *Repository) Update(
	ctx context.Context,
	userID int64,
	exposureID int64,
	input UpdateInput,
) (Exposure, error) {
	dbLensID := toDBInt8(input.LensID)

	state, err := r.queries.GetExposureUpdateState(
		ctx,
		db.GetExposureUpdateStateParams{
			CameraID: input.CameraID,
			UserID:   userID,
			LensID:   dbLensID,
			ID:       exposureID,
		},
	)
	if err != nil {
		if errors.Is(err, pgx.ErrNoRows) {
			return Exposure{}, ErrNotFound
		}

		return Exposure{}, fmt.Errorf("get exposure update state: %w", err)
	}

	if !state.CameraAvailable {
		return Exposure{}, ErrCameraNotFound
	}

	if !state.LensAvailable.Valid || !state.LensAvailable.Bool {
		return Exposure{}, ErrLensNotFound
	}

	value, err := r.queries.UpdateExposure(
		ctx,
		db.UpdateExposureParams{
			CameraID:       input.CameraID,
			LensID:         dbLensID,
			Aperture:       toDBFloat8(input.Aperture),
			ShutterSpeedUs: toDBInt8(input.ShutterSpeedUS),
			ShotAt:         toDBTimestamptz(input.ShotAt),
			Note:           toDBText(input.Note),
			ID:             exposureID,
			UserID:         userID,
		},
	)
	if err != nil {
		if errors.Is(err, pgx.ErrNoRows) {
			return Exposure{}, ErrUpdateRejected
		}

		return Exposure{}, fmt.Errorf("update exposure: %w", err)
	}

	return fromDBExposure(value), nil
}

func (r *Repository) Delete(
	ctx context.Context,
	userID int64,
	exposureID int64,
) error {
	rowsAffected, err := r.queries.DeleteExposure(
		ctx,
		db.DeleteExposureParams{
			ID:     exposureID,
			UserID: userID,
		},
	)
	if err != nil {
		return fmt.Errorf("delete exposure: %w", err)
	}

	if rowsAffected == 0 {
		return ErrNotFound
	}

	return nil
}
