package services

import (
	"context"
)

// timeFormat is the shared layout used by every service when rendering
// timestamps into *Output structs.
const timeFormat = "2006-01-02 15:04:05 -0700"

type UserService interface {
	Create(ctx context.Context, inp CreateUserInput) (*UserOutput, error)
	Get(ctx context.Context, inp GetUserInput) (*UserOutput, error)
	List(ctx context.Context, inp ListUsersInput) (*ListUsersOutput, error)
	Update(ctx context.Context, inp UpdateUserInput) (*UserOutput, error)
	Delete(ctx context.Context, inp DeleteUserInput) error
}
