package pgx

import (
	"context"

	"github.com/jackc/pgx/v5"
)

type querier interface {
	Query(ctx context.Context, dbFuncName string, sql string, namedArgs pgx.NamedArgs) (*pgx.Rows, context.CancelFunc, error)
	QueryRow(ctx context.Context, dbFuncName string, sql string, namedArgs pgx.NamedArgs) (*pgx.Row, context.CancelFunc)
}

type Connection interface {
	querier
	WithTransaction(ctx context.Context, name string, txOptions TxOptions, f func(tx ConnectionTx) error) (context.CancelFunc, error)
}

type ConnectionTx interface {
	Query(ctx context.Context, dbFuncName string, sql string, namedArgs pgx.NamedArgs) (*pgx.Rows, error)
	QueryRow(ctx context.Context, dbFuncName string, sql string, namedArgs pgx.NamedArgs) *pgx.Row
}
