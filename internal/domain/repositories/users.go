package repositories

import (
	"context"
	"database/sql"

	"github.com/google/uuid"
	"github.com/jmoiron/sqlx"
	"github.com/lib/pq"
	"github.com/louispy/gotemplate/internal/custerr"
	"github.com/louispy/gotemplate/internal/domain/models"
	"github.com/louispy/gotemplate/internal/utils"
)

type defaultUsersRepository struct {
	db *sqlx.DB
}

type UsersRepoOpts struct {
	DB *sqlx.DB
}

func NewUsersRepository(opts UsersRepoOpts) UsersRepository {
	return &defaultUsersRepository{
		db: opts.DB,
	}
}

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

const insertUserQuery = `
	INSERT INTO
		users (id, name, email, created_at, updated_at)
	VALUES
		($1, $2, $3, $4, $5)
`

func (r defaultUsersRepository) Create(ctx context.Context, user models.User) (uuid.UUID, error) {
	if user.Id == uuid.Nil {
		user.Id = uuid.New()
	}
	args := []any{user.Id, user.Name, user.Email, user.CreatedAt, user.UpdatedAt}

	var err error
	if tx := utils.SqlxTxFromCtx(ctx); tx != nil {
		_, err = tx.ExecContext(ctx, insertUserQuery, args...)
	} else {
		_, err = r.db.ExecContext(ctx, insertUserQuery, args...)
	}

	return user.Id, mapError(err)
}

const getUserQuery = `
	SELECT id, name, email, created_at, updated_at
	FROM users
	WHERE id = $1
	LIMIT 1
`

func (r defaultUsersRepository) GetById(ctx context.Context, id uuid.UUID) (*models.User, error) {
	user := models.User{}

	var err error
	if tx := utils.SqlxTxFromCtx(ctx); tx != nil {
		err = tx.GetContext(ctx, &user, getUserQuery, id)
	} else {
		err = r.db.GetContext(ctx, &user, getUserQuery, id)
	}

	if err != nil {
		return nil, mapError(err)
	}
	return &user, nil
}

const listUsersQuery = `
	SELECT id, name, email, created_at, updated_at
	FROM users
	ORDER BY created_at DESC
`

func (r defaultUsersRepository) List(ctx context.Context) ([]models.User, error) {
	users := []models.User{}

	var err error
	if tx := utils.SqlxTxFromCtx(ctx); tx != nil {
		err = tx.SelectContext(ctx, &users, listUsersQuery)
	} else {
		err = r.db.SelectContext(ctx, &users, listUsersQuery)
	}

	if err != nil {
		return nil, mapError(err)
	}
	return users, nil
}

const updateUserQuery = `
	UPDATE users
	SET name = $2, email = $3, updated_at = $4
	WHERE id = $1
`

func (r defaultUsersRepository) Update(ctx context.Context, user models.User) error {
	args := []any{user.Id, user.Name, user.Email, user.UpdatedAt}

	var (
		res sql.Result
		err error
	)
	if tx := utils.SqlxTxFromCtx(ctx); tx != nil {
		res, err = tx.ExecContext(ctx, updateUserQuery, args...)
	} else {
		res, err = r.db.ExecContext(ctx, updateUserQuery, args...)
	}
	if err != nil {
		return mapError(err)
	}

	rows, err := res.RowsAffected()
	if err != nil {
		return err
	}
	if rows == 0 {
		return custerr.ErrDataNotFound
	}
	return nil
}

const deleteUserQuery = `
	DELETE FROM users
	WHERE id = $1
`

func (r defaultUsersRepository) Delete(ctx context.Context, id uuid.UUID) error {
	var (
		res sql.Result
		err error
	)
	if tx := utils.SqlxTxFromCtx(ctx); tx != nil {
		res, err = tx.ExecContext(ctx, deleteUserQuery, id)
	} else {
		res, err = r.db.ExecContext(ctx, deleteUserQuery, id)
	}
	if err != nil {
		return mapError(err)
	}

	rows, err := res.RowsAffected()
	if err != nil {
		return err
	}
	if rows == 0 {
		return custerr.ErrDataNotFound
	}
	return nil
}
