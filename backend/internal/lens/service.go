package lens

import "context"

type LensRepository interface {
	Create(
		ctx context.Context,
		userID int64,
		manufacturer string,
		model string,
		focalLengthMm int32,
		maxAperture float64,
	) (Lens, error)

	List(
		ctx context.Context,
		userID int64,
	) ([]Lens, error)

	GetByID(
		ctx context.Context,
		userID int64,
		lensID int64,
	) (Lens, error)

	Update(
		ctx context.Context,
		userID int64,
		lensID int64,
		manufacturer string,
		model string,
		focalLengthMm int32,
		maxAperture float64,
	) (Lens, error)

	Delete(
		ctx context.Context,
		userID int64,
		lensID int64,
	) error
}

type Service struct {
	repository LensRepository
}

func NewService(repository LensRepository) *Service {
	return &Service{
		repository: repository,
	}
}

func (s *Service) Create(
	ctx context.Context,
	userID int64,
	manufacturer string,
	model string,
	focalLengthMm int32,
	maxAperture float64,
) (Lens, error) {
	return s.repository.Create(
		ctx,
		userID,
		manufacturer,
		model,
		focalLengthMm,
		maxAperture,
	)
}

func (s *Service) List(
	ctx context.Context,
	userID int64,
) ([]Lens, error) {
	return s.repository.List(ctx, userID)
}

func (s *Service) GetByID(
	ctx context.Context,
	userID int64,
	lensID int64,
) (Lens, error) {
	return s.repository.GetByID(ctx, userID, lensID)
}

func (s *Service) Update(
	ctx context.Context,
	userID int64,
	lensID int64,
	manufacturer string,
	model string,
	focalLengthMm int32,
	maxAperture float64,
) (Lens, error) {
	return s.repository.Update(
		ctx,
		userID,
		lensID,
		manufacturer,
		model,
		focalLengthMm,
		maxAperture,
	)
}

func (s *Service) Delete(
	ctx context.Context,
	userID int64,
	lensID int64,
) error {
	return s.repository.Delete(ctx, userID, lensID)
}
