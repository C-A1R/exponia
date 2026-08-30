package frame

import "context"

type FrameRepository interface {
	Create(
		ctx context.Context,
		userID int64,
		filmRollID int64,
		frameLabel *string,
		note *string,
	) (Frame, error)

	GetByID(
		ctx context.Context,
		userID int64,
		frameID int64,
	) (Frame, error)

	List(
		ctx context.Context,
		userID int64,
		filmRollID int64,
	) ([]Frame, error)

	Update(
		ctx context.Context,
		userID int64,
		frameID int64,
		frameLabel *string,
		note *string,
	) (Frame, error)

	Delete(
		ctx context.Context,
		userID int64,
		frameID int64,
	) error
}

type Service struct {
	repository FrameRepository
}

func NewService(repository FrameRepository) *Service {
	return &Service{repository: repository}
}

func (s *Service) Create(
	ctx context.Context,
	userID int64,
	filmRollID int64,
	frameLabel *string,
	note *string,
) (Frame, error) {
	return s.repository.Create(
		ctx,
		userID,
		filmRollID,
		frameLabel,
		note,
	)
}

func (s *Service) GetByID(
	ctx context.Context,
	userID int64,
	frameID int64,
) (Frame, error) {
	return s.repository.GetByID(ctx, userID, frameID)
}

func (s *Service) List(
	ctx context.Context,
	userID int64,
	filmRollID int64,
) ([]Frame, error) {
	return s.repository.List(ctx, userID, filmRollID)
}

func (s *Service) Update(
	ctx context.Context,
	userID int64,
	frameID int64,
	frameLabel *string,
	note *string,
) (Frame, error) {
	return s.repository.Update(
		ctx,
		userID,
		frameID,
		frameLabel,
		note,
	)
}

func (s *Service) Delete(
	ctx context.Context,
	userID int64,
	frameID int64,
) error {
	return s.repository.Delete(ctx, userID, frameID)
}
