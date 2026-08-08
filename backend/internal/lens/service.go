package lens

import "context"

type LensRepository interface {
	Create(
		ctx context.Context,
		manufacturer string,
		model string,
		focalLengthMm int32,
		maxAperture float64,
	) (Lens, error)

	List(ctx context.Context) ([]Lens, error)

	GetByID(
		ctx context.Context,
		id int64,
	) (Lens, error)

	Update(
		ctx context.Context,
		id int64,
		manufacturer string,
		model string,
		focalLengthMm int32,
		maxAperture float64,
	) (Lens, error)

	Delete(
		ctx context.Context,
		id int64,
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
	manufacturer string,
	model string,
	focalLengthMm int32,
	maxAperture float64,
) (Lens, error) {
	return s.repository.Create(
		ctx,
		manufacturer,
		model,
		focalLengthMm,
		maxAperture,
	)
}

func (s *Service) List(ctx context.Context) ([]Lens, error) {
	return s.repository.List(ctx)
}

func (s *Service) GetByID(
	ctx context.Context,
	id int64,
) (Lens, error) {
	return s.repository.GetByID(ctx, id)
}

func (s *Service) Update(
	ctx context.Context,
	id int64,
	manufacturer string,
	model string,
	focalLengthMm int32,
	maxAperture float64,
) (Lens, error) {
	return s.repository.Update(
		ctx,
		id,
		manufacturer,
		model,
		focalLengthMm,
		maxAperture,
	)
}

func (s *Service) Delete(
	ctx context.Context,
	id int64,
) error {
	return s.repository.Delete(ctx, id)
}
