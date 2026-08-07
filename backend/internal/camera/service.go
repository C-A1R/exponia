package camera

import "context"

type CameraRepository interface {
	CreateCamera(
		ctx context.Context,
		manufacturer string,
		model string,
	) (Camera, error)

	ListCameras(ctx context.Context) ([]Camera, error)

	GetCameraByID(
		ctx context.Context,
		id int64,
	) (Camera, error)

	UpdateCamera(
		ctx context.Context,
		id int64,
		manufacturer string,
		model string,
	) (Camera, error)

	DeleteCamera(
		ctx context.Context,
		id int64,
	) error
}

type Service struct {
	repository CameraRepository
}

func NewService(repository CameraRepository) *Service {
	return &Service{
		repository: repository,
	}
}

func (s *Service) CreateCamera(
	ctx context.Context,
	manufacturer string,
	model string,
) (Camera, error) {
	return s.repository.CreateCamera(ctx, manufacturer, model)
}

func (s *Service) ListCameras(ctx context.Context) ([]Camera, error) {
	return s.repository.ListCameras(ctx)
}

func (s *Service) GetCameraByID(ctx context.Context, id int64) (Camera, error) {
	return s.repository.GetCameraByID(ctx, id)
}

func (s *Service) UpdateCamera(
	ctx context.Context,
	id int64,
	manufacturer string,
	model string,
) (Camera, error) {
	return s.repository.UpdateCamera(ctx, id, manufacturer, model)
}

func (s *Service) DeleteCamera(ctx context.Context, id int64) error {
	return s.repository.DeleteCamera(ctx, id)
}
