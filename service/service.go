package service

import (
	"context"

	"github.com/pkg/errors"
	"github.com/torwig/user-service/entities"
)

type UserRepository interface {
	Create(ctx context.Context, params entities.CreateUserParams) (entities.User, error)
	Get(ctx context.Context, id int64) (entities.User, error)
	Update(ctx context.Context, id int64, params entities.UpdateUserParams) (entities.User, error)
	Delete(ctx context.Context, id int64) error
}

type Service struct {
	userRepo UserRepository
}

func New(userRepo UserRepository) *Service {
	return &Service{userRepo: userRepo}
}

func (s *Service) CreateUser(ctx context.Context, params entities.CreateUserParams) (entities.User, error) {
	user, err := s.userRepo.Create(ctx, params)
	if err != nil {
		return user, errors.Wrap(err, "failed to create user in repository")
	}

	return user, nil
}

func (s *Service) GetUser(ctx context.Context, id int64) (entities.User, error) {
	user, err := s.userRepo.Get(ctx, id)
	if err != nil {
		return entities.User{}, errors.Wrap(err, "failed to get user from repository")
	}

	if user.IsDeleted() {
		return entities.User{}, entities.ErrUserNotFound
	}

	return user, nil
}

func (s *Service) UpdateUser(ctx context.Context, id int64, params entities.UpdateUserParams) (entities.User, error) {
	existingUser, err := s.userRepo.Get(ctx, id)
	if err != nil {
		return entities.User{}, errors.Wrap(err, "failed to get user from repository")
	}

	if existingUser.IsDeleted() {
		return entities.User{}, entities.ErrUserNotFound
	}

	updatedUser, err := s.userRepo.Update(ctx, id, params)
	if err != nil {
		return entities.User{}, errors.Wrap(err, "failed to update user in repository")
	}

	return updatedUser, nil
}

func (s *Service) DeleteUser(ctx context.Context, id int64) error {
	_, err := s.userRepo.Get(ctx, id)
	if err != nil {
		return errors.Wrap(err, "failed to get user from repository")
	}

	if err := s.userRepo.Delete(ctx, id); err != nil {
		return errors.Wrap(err, "failed to delete user from repository")
	}

	return nil
}
