package exposure

import (
	"context"
	"errors"
	"math"
)

var (
	ErrInvalidCameraID     = errors.New("invalid camera id")
	ErrInvalidLensID       = errors.New("invalid lens id")
	ErrInvalidAperture     = errors.New("invalid aperture")
	ErrInvalidShutterSpeed = errors.New("invalid shutter speed")
)

type ExposureRepository interface {
	Create(
		ctx context.Context,
		userID int64,
		frameID int64,
		input CreateInput,
	) (Exposure, error)

	GetByID(
		ctx context.Context,
		userID int64,
		exposureID int64,
	) (Exposure, error)

	List(
		ctx context.Context,
		userID int64,
		frameID int64,
	) ([]Exposure, error)

	Update(
		ctx context.Context,
		userID int64,
		exposureID int64,
		input UpdateInput,
	) (Exposure, error)

	Delete(
		ctx context.Context,
		userID int64,
		exposureID int64,
	) error
}

type Service struct {
	repository ExposureRepository
}

func NewService(repository ExposureRepository) *Service {
	return &Service{repository: repository}
}

func validateOptionalValues(
	lensID *int64,
	aperture *float64,
	shutterSpeedUS *int64,
) error {
	if lensID != nil && *lensID <= 0 {
		return ErrInvalidLensID
	}

	if aperture != nil &&
		(*aperture <= 0 || math.IsNaN(*aperture) || math.IsInf(*aperture, 0)) {
		return ErrInvalidAperture
	}

	if shutterSpeedUS != nil && *shutterSpeedUS <= 0 {
		return ErrInvalidShutterSpeed
	}

	return nil
}

func (s *Service) Create(
	ctx context.Context,
	userID int64,
	frameID int64,
	input CreateInput,
) (Exposure, error) {
	if err := validateOptionalValues(
		input.LensID,
		input.Aperture,
		input.ShutterSpeedUS,
	); err != nil {
		return Exposure{}, err
	}

	return s.repository.Create(ctx, userID, frameID, input)
}

func (s *Service) GetByID(
	ctx context.Context,
	userID int64,
	exposureID int64,
) (Exposure, error) {
	return s.repository.GetByID(ctx, userID, exposureID)
}

func (s *Service) List(
	ctx context.Context,
	userID int64,
	frameID int64,
) ([]Exposure, error) {
	return s.repository.List(ctx, userID, frameID)
}

func (s *Service) Update(
	ctx context.Context,
	userID int64,
	exposureID int64,
	input UpdateInput,
) (Exposure, error) {
	if input.CameraID <= 0 {
		return Exposure{}, ErrInvalidCameraID
	}

	if err := validateOptionalValues(
		input.LensID,
		input.Aperture,
		input.ShutterSpeedUS,
	); err != nil {
		return Exposure{}, err
	}

	return s.repository.Update(ctx, userID, exposureID, input)
}

func (s *Service) Delete(
	ctx context.Context,
	userID int64,
	exposureID int64,
) error {
	return s.repository.Delete(ctx, userID, exposureID)
}
