package database

import (
	"context"
	"errors"

	"github.com/jmoiron/sqlx"
	"github.com/louispy/gotemplate/internal/constants"
	"github.com/louispy/gotemplate/internal/utils"
)

type TxManager interface {
	Begin(ctx context.Context) (context.Context, error)
	Commit(ctx context.Context) error
	Rollback(ctx context.Context) error
}

type defaultTxManager struct {
	db *sqlx.DB
}

type TxManagerOpts struct {
	DB *sqlx.DB
}

func NewTxManager(opts TxManagerOpts) TxManager {
	return &defaultTxManager{
		db: opts.DB,
	}
}

func (m defaultTxManager) Begin(ctx context.Context) (context.Context, error) {
	tx, err := m.db.Beginx()
	if err != nil {
		return nil, err
	}
	ctx = context.WithValue(ctx, constants.SqlxTxCtx, tx)

	return ctx, nil
}

func (m defaultTxManager) Commit(ctx context.Context) error {
	tx := utils.SqlxTxFromCtx(ctx)
	if tx == nil {
		return errors.New("invalid tx")
	}
	return tx.Commit()
}

func (m defaultTxManager) Rollback(ctx context.Context) error {
	tx := utils.SqlxTxFromCtx(ctx)
	if tx == nil {
		return errors.New("invalid tx")
	}

	return tx.Rollback()
}
