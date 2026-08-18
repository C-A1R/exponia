package user

import (
	"context"
)

type UserRepository interface {
	GetByAuthIdentity(
		ctx context.Context,
		authIssuer string,
		authSubject string,
	) (User, error)

	FindOrCreate(
		ctx context.Context,
		email string,
		displayName string,
		authIssuer string,
		authSubject string,
	) (User, error)
}

type Service struct {
	repository UserRepository
}

func NewService(repository UserRepository) *Service {
	return &Service{
		repository: repository,
	}
}

func (s *Service) GetByAuthIdentity(
	ctx context.Context,
	authIssuer string,
	authSubject string,
) (User, error) {
	return s.repository.GetByAuthIdentity(
		ctx,
		authIssuer,
		authSubject,
	)
}

func (s *Service) FindOrCreate(
	ctx context.Context,
	email string,
	displayName string,
	authIssuer string,
	authSubject string,
) (User, error) {
	return s.repository.FindOrCreate(
		ctx,
		email,
		displayName,
		authIssuer,
		authSubject,
	)
}
