package pgx

import (
	"context"
	"time"

	"github.com/jackc/pgx/v5"
)

type Transaction struct {
	tx *pgx.Tx
	c  *ConnectionPool
}

func (t *Transaction) QueryRow(
	ctx context.Context,
	dbFuncName string,
	sql string,
	namedArgs pgx.NamedArgs,
) *pgx.Row {
	start := time.Now()
	r := (*t.tx).QueryRow(ctx, sql, namedArgs)
	diff := time.Since(start)

	if t.c.metrics.qm != nil {
		t.c.metrics.qm.ObserveQueryDurationHistogram(diff.Seconds(), dbFuncName)
		t.c.metrics.qm.IncQueryCounter(querySuccess, dbFuncName)
	}

	return &r
}

func (t *Transaction) Query(
	ctx context.Context,
	dbFuncName string,
	sql string,
	namedArgs pgx.NamedArgs,
) (*pgx.Rows, error) {
	start := time.Now()
	r, err := (*t.tx).Query(ctx, sql, namedArgs)

	if err != nil {
		if t.c.metrics.qm != nil {
			t.c.metrics.qm.IncQueryCounter(queryError, dbFuncName)
		}

		return nil, err
	}
	err = r.Err()
	if err != nil {
		if t.c.metrics.qm != nil {
			t.c.metrics.qm.IncQueryCounter(queryError, dbFuncName)
		}

		return nil, err
	}

	diff := time.Since(start)

	if t.c.metrics.qm != nil {
		t.c.metrics.qm.ObserveQueryDurationHistogram(diff.Seconds(), dbFuncName)
		t.c.metrics.qm.IncQueryCounter(querySuccess, dbFuncName)
	}

	return &r, nil
}
