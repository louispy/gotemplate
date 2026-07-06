package repositories

import (
	"context"

	"github.com/google/uuid"
	"github.com/louispy/gotemplate/internal/domain/models"
)

type UsersRepository interface {
	Create(ctx context.Context, user models.User) (uuid.UUID, error)
	GetById(ctx context.Context, id uuid.UUID) (*models.User, error)
	List(ctx context.Context) ([]models.User, error)
	Update(ctx context.Context, user models.User) error
	Delete(ctx context.Context, id uuid.UUID) error
}
