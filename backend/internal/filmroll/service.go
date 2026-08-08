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
		filmStockID int64,
		formatID int64,
	) (FilmRoll, error)

	GetByID(
		ctx context.Context,
		id int64,
	) (FilmRoll, error)

	List(
		ctx context.Context,
	) ([]FilmRoll, error)

	UpdateStatus(
		ctx context.Context,
		id int64,
		status FilmRollStatus,
	) (FilmRoll, error)

	UpdateExposureISO(
		ctx context.Context,
		id int64,
		exposureISO int32,
	) (FilmRoll, error)

	UpdateCamera(
		ctx context.Context,
		id int64,
		cameraID *int64,
	) (FilmRoll, error)

	Delete(
		ctx context.Context,
		id int64,
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
	filmStockID int64,
	formatID int64,
) (FilmRoll, error) {
	return s.repository.Create(ctx, filmStockID, formatID)
}

func (s *Service) GetByID(
	ctx context.Context,
	id int64,
) (FilmRoll, error) {
	return s.repository.GetByID(ctx, id)
}

func (s *Service) List(
	ctx context.Context,
) ([]FilmRoll, error) {
	return s.repository.List(ctx)
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
	id int64,
	status FilmRollStatus,
) (FilmRoll, error) {
	if !isValidStatus(status) {
		return FilmRoll{}, ErrInvalidStatus
	}

	return s.repository.UpdateStatus(ctx, id, status)
}

func (s *Service) UpdateExposureISO(
	ctx context.Context,
	id int64,
	exposureISO int32,
) (FilmRoll, error) {
	if exposureISO <= 0 {
		return FilmRoll{}, ErrInvalidExposureISO
	}

	return s.repository.UpdateExposureISO(ctx, id, exposureISO)
}

func (s *Service) UpdateCamera(
	ctx context.Context,
	id int64,
	cameraID *int64,
) (FilmRoll, error) {
	return s.repository.UpdateCamera(ctx, id, cameraID)
}

func (s *Service) Delete(
	ctx context.Context,
	id int64,
) error {
	return s.repository.Delete(ctx, id)
}
