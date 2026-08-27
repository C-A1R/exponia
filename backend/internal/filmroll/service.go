package filmroll

import (
	"context"
	"errors"
)

var ErrInvalidStatus = errors.New("invalid film roll status")
var ErrInvalidExposureISO = errors.New("invalid exposure ISO")

type FilmRollRepository interface {
	Create(
		ctx context.Context,
		userID int64,
		filmStockID int64,
		formatID int64,
	) (FilmRoll, error)

	GetByID(
		ctx context.Context,
		userID int64,
		filmRollID int64,
	) (FilmRoll, error)

	List(
		ctx context.Context,
		userID int64,
	) ([]FilmRoll, error)

	UpdateStatus(
		ctx context.Context,
		userID int64,
		filmRollID int64,
		status FilmRollStatus,
	) (FilmRoll, error)

	UpdateExposureISO(
		ctx context.Context,
		userID int64,
		filmRollID int64,
		exposureISO int32,
	) (FilmRoll, error)

	UpdateCamera(
		ctx context.Context,
		userID int64,
		filmRollID int64,
		cameraID *int64,
	) (FilmRoll, error)

	Delete(
		ctx context.Context,
		userID int64,
		filmRollID int64,
	) error
}

type Service struct {
	repository FilmRollRepository
}

func NewService(repository FilmRollRepository) *Service {
	return &Service{
		repository: repository,
	}
}

func (s *Service) Create(
	ctx context.Context,
	userID int64,
	filmStockID int64,
	formatID int64,
) (FilmRoll, error) {
	return s.repository.Create(ctx, userID, filmStockID, formatID)
}

func (s *Service) GetByID(
	ctx context.Context,
	userID int64,
	filmRollID int64,
) (FilmRoll, error) {
	return s.repository.GetByID(ctx, userID, filmRollID)
}

func (s *Service) List(
	ctx context.Context,
	userID int64,
) ([]FilmRoll, error) {
	return s.repository.List(ctx, userID)
}

func isValidStatus(status FilmRollStatus) bool {
	switch status {
	case FilmRollStatusUnused,
		FilmRollStatusReady,
		FilmRollStatusInUse,
		FilmRollStatusExposed,
		FilmRollStatusDeveloped:
		return true
	default:
		return false
	}
}

func (s *Service) UpdateStatus(
	ctx context.Context,
	userID int64,
	filmRollID int64,
	status FilmRollStatus,
) (FilmRoll, error) {
	if !isValidStatus(status) {
		return FilmRoll{}, ErrInvalidStatus
	}

	return s.repository.UpdateStatus(ctx, userID, filmRollID, status)
}

func (s *Service) UpdateExposureISO(
	ctx context.Context,
	userID int64,
	filmRollID int64,
	exposureISO int32,
) (FilmRoll, error) {
	if exposureISO <= 0 {
		return FilmRoll{}, ErrInvalidExposureISO
	}

	return s.repository.UpdateExposureISO(
		ctx,
		userID,
		filmRollID,
		exposureISO,
	)
}

func (s *Service) UpdateCamera(
	ctx context.Context,
	userID int64,
	filmRollID int64,
	cameraID *int64,
) (FilmRoll, error) {
	return s.repository.UpdateCamera(ctx, userID, filmRollID, cameraID)
}

func (s *Service) Delete(
	ctx context.Context,
	userID int64,
	filmRollID int64,
) error {
	return s.repository.Delete(ctx, userID, filmRollID)
}
