package user

import (
	"context"
	"errors"
	"fmt"

	db "github.com/C-A1R/exponia/backend/internal/database/generated"
	"github.com/jackc/pgx/v5"
	"github.com/jackc/pgx/v5/pgxpool"
)

var ErrNotFound = errors.New("user not found")

type Repository struct {
	queries *db.Queries
}

func NewRepository(pool *pgxpool.Pool) *Repository {
	return &Repository{
		queries: db.New(pool),
	}
}

func fromDBUser(value db.User) User {
	return User{
		ID:          value.ID,
		Email:       value.Email,
		DisplayName: value.DisplayName,
		AuthIssuer:  value.AuthIssuer,
		AuthSubject: value.AuthSubject,
		CreatedAt:   value.CreatedAt.Time,
	}
}

func (r *Repository) Create(
	ctx context.Context,
	email string,
	displayName string,
	authIssuer string,
	authSubject string,
) (User, error) {
	value, err := r.queries.CreateUser(
		ctx,
		db.CreateUserParams{
			Email:       email,
			DisplayName: displayName,
			AuthIssuer:  authIssuer,
			AuthSubject: authSubject,
		},
	)
	if err != nil {
		return User{}, fmt.Errorf("create user: %w", err)
	}

	return fromDBUser(value), nil
}

func (r *Repository) FindOrCreate(
	ctx context.Context,
	email string,
	displayName string,
	authIssuer string,
	authSubject string,
) (User, error) {
	value, err := r.queries.FindOrCreateUser(
		ctx,
		db.FindOrCreateUserParams{
			Email:       email,
			DisplayName: displayName,
			AuthIssuer:  authIssuer,
			AuthSubject: authSubject,
		},
	)
	if err != nil {
		return User{}, fmt.Errorf("find or create user: %w", err)
	}

	return fromDBUser(value), nil
}

func (r *Repository) GetByID(
	ctx context.Context,
	id int64,
) (User, error) {
	value, err := r.queries.GetUserByID(ctx, id)
	if err != nil {
		if errors.Is(err, pgx.ErrNoRows) {
			return User{}, ErrNotFound
		}

		return User{}, fmt.Errorf("get user by id: %w", err)
	}

	return fromDBUser(value), nil
}

func (r *Repository) GetByAuthIdentity(
	ctx context.Context,
	authIssuer string,
	authSubject string,
) (User, error) {
	value, err := r.queries.GetUserByAuthIdentity(
		ctx,
		db.GetUserByAuthIdentityParams{
			AuthIssuer:  authIssuer,
			AuthSubject: authSubject,
		},
	)
	if err != nil {
		if errors.Is(err, pgx.ErrNoRows) {
			return User{}, ErrNotFound
		}

		return User{}, fmt.Errorf(
			"get user by auth identity: %w",
			err,
		)
	}

	return fromDBUser(value), nil
}
