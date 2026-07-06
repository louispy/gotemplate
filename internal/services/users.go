package services

import (
	"context"
	"log"
	"time"

	"github.com/google/uuid"
	"github.com/louispy/gotemplate/internal/custerr"
	"github.com/louispy/gotemplate/internal/domain/models"
	"github.com/louispy/gotemplate/internal/domain/repositories"
)

type userService struct {
	usersRepo repositories.UsersRepository
}

type UserServiceOpts struct {
	UsersRepo repositories.UsersRepository
}

func NewUserService(opts UserServiceOpts) UserService {
	return &userService{
		usersRepo: opts.UsersRepo,
	}
}

func toOutput(user models.User) UserOutput {
	return UserOutput{
		Id:        user.Id.String(),
		Name:      user.Name,
		Email:     user.Email,
		CreatedAt: user.CreatedAt.Format(timeFormat),
		UpdatedAt: user.UpdatedAt.Format(timeFormat),
	}
}

func (s userService) Create(ctx context.Context, inp CreateUserInput) (*UserOutput, error) {
	now := time.Now().UTC()
	user := models.User{
		Id:        uuid.New(),
		Name:      inp.Name,
		Email:     inp.Email,
		CreatedAt: now,
		UpdatedAt: now,
	}

	if _, err := s.usersRepo.Create(ctx, user); err != nil {
		if err != custerr.ErrDuplicate {
			log.Printf("Error creating user: %v\n", err.Error())
		}
		return nil, err
	}

	out := toOutput(user)
	return &out, nil
}

func (s userService) Get(ctx context.Context, inp GetUserInput) (*UserOutput, error) {
	user, err := s.usersRepo.GetById(ctx, inp.Id)
	if err != nil {
		if err != custerr.ErrDataNotFound {
			log.Printf("Error retrieving user: %v\n", err.Error())
		}
		return nil, err
	}

	out := toOutput(*user)
	return &out, nil
}

func (s userService) List(ctx context.Context, inp ListUsersInput) (*ListUsersOutput, error) {
	users, err := s.usersRepo.List(ctx)
	if err != nil {
		log.Printf("Error listing users: %v\n", err.Error())
		return nil, err
	}

	out := make([]UserOutput, 0, len(users))
	for _, user := range users {
		out = append(out, toOutput(user))
	}
	return &ListUsersOutput{Users: out}, nil
}

func (s userService) Update(ctx context.Context, inp UpdateUserInput) (*UserOutput, error) {
	existing, err := s.usersRepo.GetById(ctx, inp.Id)
	if err != nil {
		if err != custerr.ErrDataNotFound {
			log.Printf("Error retrieving user: %v\n", err.Error())
		}
		return nil, err
	}

	existing.Name = inp.Name
	existing.Email = inp.Email
	existing.UpdatedAt = time.Now().UTC()

	if err := s.usersRepo.Update(ctx, *existing); err != nil {
		if err != custerr.ErrDataNotFound && err != custerr.ErrDuplicate {
			log.Printf("Error updating user: %v\n", err.Error())
		}
		return nil, err
	}

	out := toOutput(*existing)
	return &out, nil
}

func (s userService) Delete(ctx context.Context, inp DeleteUserInput) error {
	if err := s.usersRepo.Delete(ctx, inp.Id); err != nil {
		if err != custerr.ErrDataNotFound {
			log.Printf("Error deleting user: %v\n", err.Error())
		}
		return err
	}
	return nil
}
