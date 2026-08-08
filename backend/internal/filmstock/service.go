package filmstock

import "context"

type FilmStockRepository interface {
	List(ctx context.Context) ([]FilmStock, error)
	GetByID(ctx context.Context, id int64) (FilmStock, error)
}

type Service struct {
	repository FilmStockRepository
}

func NewService(repository FilmStockRepository) *Service {
	return &Service{
		repository: repository,
	}
}

func (s *Service) List(ctx context.Context) ([]FilmStock, error) {
	return s.repository.List(ctx)
}

func (s *Service) GetByID(ctx context.Context, id int64) (FilmStock, error) {
	return s.repository.GetByID(ctx, id)
}
