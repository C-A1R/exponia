package camera

import "context"

type CameraRepository interface {
	CreateCamera(
		ctx context.Context,
		userID int64,
		manufacturer string,
		model string,
	) (Camera, error)

	ListCameras(
		ctx context.Context,
		userID int64,
	) ([]Camera, error)

	GetCameraByID(
		ctx context.Context,
		userID int64,
		cameraID int64,
	) (Camera, error)

	UpdateCamera(
		ctx context.Context,
		userID int64,
		cameraID int64,
		manufacturer string,
		model string,
	) (Camera, error)

	DeleteCamera(
		ctx context.Context,
		userID int64,
		cameraID int64,
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
	userID int64,
	manufacturer string,
	model string,
) (Camera, error) {
	return s.repository.CreateCamera(ctx, userID, manufacturer, model)
}

func (s *Service) ListCameras(
	ctx context.Context,
	userID int64,
) ([]Camera, error) {
	return s.repository.ListCameras(ctx, userID)
}

func (s *Service) GetCameraByID(
	ctx context.Context,
	userID int64,
	cameraID int64,
) (Camera, error) {
	return s.repository.GetCameraByID(ctx, userID, cameraID)
}

func (s *Service) UpdateCamera(
	ctx context.Context,
	userID int64,
	cameraID int64,
	manufacturer string,
	model string,
) (Camera, error) {
	return s.repository.UpdateCamera(
		ctx,
		userID,
		cameraID,
		manufacturer,
		model,
	)
}

func (s *Service) DeleteCamera(
	ctx context.Context,
	userID int64,
	cameraID int64,
) error {
	return s.repository.DeleteCamera(ctx, userID, cameraID)
}
