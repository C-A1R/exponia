package frame

import (
	"context"
	"errors"
	"fmt"

	db "github.com/C-A1R/exponia/backend/internal/database/generated"
	"github.com/jackc/pgx/v5"
	"github.com/jackc/pgx/v5/pgtype"
	"github.com/jackc/pgx/v5/pgxpool"
)

var (
	ErrNotFound         = errors.New("frame not found")
	ErrFilmRollNotFound = errors.New("film roll not found")
)

type Repository struct {
	queries *db.Queries
}

func NewRepository(pool *pgxpool.Pool) *Repository {
	return &Repository{
		queries: db.New(pool),
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

func fromDBText(value pgtype.Text) *string {
	if !value.Valid {
		return nil
	}

	result := value.String

	return &result
}

func fromDBFrame(value db.Frame) Frame {
	return Frame{
		ID:         value.ID,
		FilmRollID: value.FilmRollID,
		FrameIndex: value.FrameIndex,
		FrameLabel: fromDBText(value.FrameLabel),
		Note:       fromDBText(value.Note),
		CreatedAt:  value.CreatedAt.Time,
	}
}

func (r *Repository) Create(
	ctx context.Context,
	userID int64,
	filmRollID int64,
	frameLabel *string,
	note *string,
) (Frame, error) {
	value, err := r.queries.CreateFrame(
		ctx,
		db.CreateFrameParams{
			FrameLabel: toDBText(frameLabel),
			Note:       toDBText(note),
			FilmRollID: filmRollID,
			UserID:     userID,
		},
	)
	if err != nil {
		if errors.Is(err, pgx.ErrNoRows) {
			return Frame{}, ErrFilmRollNotFound
		}

		return Frame{}, fmt.Errorf("create frame: %w", err)
	}

	return fromDBFrame(value), nil
}

func (r *Repository) GetByID(
	ctx context.Context,
	userID int64,
	frameID int64,
) (Frame, error) {
	value, err := r.queries.GetFrameByID(
		ctx,
		db.GetFrameByIDParams{
			ID:     frameID,
			UserID: userID,
		},
	)
	if err != nil {
		if errors.Is(err, pgx.ErrNoRows) {
			return Frame{}, ErrNotFound
		}

		return Frame{}, fmt.Errorf("get frame by id: %w", err)
	}

	return fromDBFrame(value), nil
}

func (r *Repository) List(
	ctx context.Context,
	userID int64,
	filmRollID int64,
) ([]Frame, error) {
	_, err := r.queries.GetOwnedFilmRollID(
		ctx,
		db.GetOwnedFilmRollIDParams{
			FilmRollID: filmRollID,
			UserID:     userID,
		},
	)
	if err != nil {
		if errors.Is(err, pgx.ErrNoRows) {
			return nil, ErrFilmRollNotFound
		}

		return nil, fmt.Errorf("get film roll for frame list: %w", err)
	}

	values, err := r.queries.ListFrames(
		ctx,
		db.ListFramesParams{
			FilmRollID: filmRollID,
			UserID:     userID,
		},
	)
	if err != nil {
		return nil, fmt.Errorf("list frames: %w", err)
	}

	frames := make([]Frame, 0, len(values))
	for _, value := range values {
		frames = append(frames, fromDBFrame(value))
	}

	return frames, nil
}

func (r *Repository) Update(
	ctx context.Context,
	userID int64,
	frameID int64,
	frameLabel *string,
	note *string,
) (Frame, error) {
	value, err := r.queries.UpdateFrame(
		ctx,
		db.UpdateFrameParams{
			FrameLabel: toDBText(frameLabel),
			Note:       toDBText(note),
			ID:         frameID,
			UserID:     userID,
		},
	)
	if err != nil {
		if errors.Is(err, pgx.ErrNoRows) {
			return Frame{}, ErrNotFound
		}

		return Frame{}, fmt.Errorf("update frame: %w", err)
	}

	return fromDBFrame(value), nil
}

func (r *Repository) Delete(
	ctx context.Context,
	userID int64,
	frameID int64,
) error {
	rowsAffected, err := r.queries.DeleteFrame(
		ctx,
		db.DeleteFrameParams{
			ID:     frameID,
			UserID: userID,
		},
	)
	if err != nil {
		return fmt.Errorf("delete frame: %w", err)
	}

	if rowsAffected == 0 {
		return ErrNotFound
	}

	return nil
}
