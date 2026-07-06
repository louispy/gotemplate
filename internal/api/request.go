package api

import (
	"errors"
	"strings"
)

type CreateUserRequest struct {
	Name  string `json:"name"`
	Email string `json:"email"`
}

func (r CreateUserRequest) Validate() error {
	return validateNameAndEmail(r.Name, r.Email)
}

type UpdateUserRequest struct {
	Name  string `json:"name"`
	Email string `json:"email"`
}

func (r UpdateUserRequest) Validate() error {
	return validateNameAndEmail(r.Name, r.Email)
}

func validateNameAndEmail(name, email string) error {
	if strings.TrimSpace(name) == "" {
		return errors.New("name is required")
	}
	if !strings.Contains(email, "@") {
		return errors.New("valid email is required")
	}
	return nil
}
