package utils

import (
	"context"

	"github.com/jmoiron/sqlx"
	"github.com/louispy/gotemplate/internal/constants"
)

func SqlxTxFromCtx(ctx context.Context) *sqlx.Tx {
	sqlxTx, ok := ctx.Value(constants.SqlxTxCtx).(*sqlx.Tx)
	if !ok {
		return nil
	}

	return sqlxTx
}
