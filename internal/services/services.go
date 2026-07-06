package services

import (
	"context"
)

type UserService interface {
	Create(ctx context.Context, inp CreateUserInput) (*UserOutput, error)
	Get(ctx context.Context, inp GetUserInput) (*UserOutput, error)
	List(ctx context.Context, inp ListUsersInput) (*ListUsersOutput, error)
	Update(ctx context.Context, inp UpdateUserInput) (*UserOutput, error)
	Delete(ctx context.Context, inp DeleteUserInput) error
}
