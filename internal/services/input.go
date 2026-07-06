package services

import (
	"github.com/google/uuid"
)

type CreateUserInput struct {
	Name  string
	Email string
}

type GetUserInput struct {
	Id uuid.UUID
}

type ListUsersInput struct{}

type UpdateUserInput struct {
	Id    uuid.UUID
	Name  string
	Email string
}

type DeleteUserInput struct {
	Id uuid.UUID
}
