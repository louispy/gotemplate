package repositories

import (
	"context"
	"database/sql"

	"github.com/google/uuid"
	"github.com/lib/pq"
	"github.com/louispy/gotemplate/internal/custerr"
	"github.com/louispy/gotemplate/internal/domain/models"
)

// mapError normalizes driver-specific errors into the shared custerr set.
// It is shared across every repository, so it lives here rather than in a
// per-resource file.
func mapError(err error) error {
	if err == nil {
		return nil
	}
	if err == sql.ErrNoRows {
		return custerr.ErrDataNotFound
	}
	if pqErr, ok := err.(*pq.Error); ok && pqErr.Code == "23505" {
		return custerr.ErrDuplicate
	}
	return err
}

type UsersRepository interface {
	Create(ctx context.Context, user models.User) (uuid.UUID, error)
	GetById(ctx context.Context, id uuid.UUID) (*models.User, error)
	List(ctx context.Context) ([]models.User, error)
	Update(ctx context.Context, user models.User) error
	Delete(ctx context.Context, id uuid.UUID) error
}
