package pgx

import (
	"context"
	"errors"
	"fmt"
	"time"

	"github.com/jackc/pgx/v5"
	"github.com/jackc/pgx/v5/pgxpool"
	"github.com/jamm3e3333/whalebone-go-test-project/pkg/logger"
)

type NamedArgs = pgx.NamedArgs

var ErrNoRows = pgx.ErrNoRows

type TxOptions struct {
	IsoLevel       TxIsoLevel
	AccessMode     TxAccessMode
	DeferrableMode TxDeferrableMode

	// BeginQuery is for custom tx options
	BeginQuery string
}

type TxIsoLevel = string

const (
	Serializable    TxIsoLevel = "serializable"
	RepeatableRead  TxIsoLevel = "repeatable read"
	ReadCommitted   TxIsoLevel = "read committed"
	ReadUncommitted TxIsoLevel = "read uncommitted"
)

type TxAccessMode string

const (
	ReadWrite TxAccessMode = "read write"
	ReadOnly  TxAccessMode = "read only"
)

type TxDeferrableMode string

const (
	Deferrable    TxDeferrableMode = "deferrable"
	NotDeferrable TxDeferrableMode = "not deferrable"
)

const (
	queryError          = "error"
	querySuccess        = "success"
	transactionCommit   = "commit"
	transactionRollback = "rollback"
)

type ConnectionPool struct {
	pool         *pgxpool.Pool
	metrics      MonitoringMetrics
	log          logger.Logger
	queryTimeout time.Duration
}

type RegisterMetricsOptions struct {
	Qm QueryMetrics
	Tm TransactionMetrics
}

func (c *ConnectionPool) RegisterMetrics(rmo RegisterMetricsOptions) {
	if rmo.Qm != nil {
		c.metrics.qm = rmo.Qm
	}

	if rmo.Tm != nil {
		c.metrics.tm = rmo.Tm
	}
}

func afterConnWithMet(cm ConnectionMetrics) func(ctx context.Context, connCfg *pgx.Conn) error {
	return func(ctx context.Context, connCfg *pgx.Conn) error {
		cm.IncDbConnGauge()
		return nil
	}
}

func beforeCloseWithMet(cm ConnectionMetrics) func(*pgx.Conn) {
	return func(*pgx.Conn) {
		cm.DecDbConnGauge()
	}
}

func NewConnectionPool(
	ctx context.Context,
	cfg Config,
	log logger.Logger,
	cm ConnectionMetrics,
) (*ConnectionPool, error) {
	connConfig, err := pgx.ParseConfig(cfg.ConnectionURL)
	if err != nil {
		return nil, fmt.Errorf("conn ParseConfig %s : %v", cfg.ConnectionURL, err)
	}

	connConfig.DefaultQueryExecMode = pgx.QueryExecModeCacheStatement

	poolCfg, err := pgxpool.ParseConfig(cfg.ConnectionURL)
	if err != nil {
		return nil, fmt.Errorf("conn pool ParseConfig %s : %v", cfg.ConnectionURL, err)
	}
	poolCfg.ConnConfig = connConfig
	poolCfg.MaxConnLifetime = cfg.MaxConnLifetime
	poolCfg.MaxConnIdleTime = cfg.MaxConnIdleTime
	poolCfg.MaxConns = cfg.DefaultMaxConns
	poolCfg.MinConns = cfg.DefaultMinConns
	poolCfg.HealthCheckPeriod = cfg.HealthCheckPeriod

	poolCfg.AfterConnect = afterConnWithMet(cm)
	poolCfg.BeforeClose = beforeCloseWithMet(cm)

	connPool, err := pgxpool.NewWithConfig(ctx, poolCfg)
	if err != nil {
		return nil, fmt.Errorf("unable to connect to database %s : %v", cfg.ConnectionURL, err)
	}

	c := ConnectionPool{
		pool:         connPool,
		log:          log,
		queryTimeout: cfg.QueryTimeout,
	}

	return &c, nil
}

func (c *ConnectionPool) WithTransaction(ctx context.Context, name string, txOptions TxOptions, f func(tx ConnectionTx) error) (context.CancelFunc, error) {
	start := time.Now()

	ctx, cancel := context.WithTimeout(ctx, c.queryTimeout)
	tx, err := c.pool.BeginTx(ctx, pgx.TxOptions{})

	if err != nil {
		c.log.ErrorWithMetadata("unable to start transaction", map[string]any{
			"error": err.Error(),
			"name":  name,
		})
		return cancel, err
	}

	tErr := f(&Transaction{
		tx: &tx,
		c:  c,
	})

	if tErr != nil {
		if c.metrics.tm != nil {
			c.metrics.tm.IncTransactionCounter(transactionRollback, name)
		}
		c.log.ErrorWithMetadata("unexpected error during rollback", map[string]any{
			"error": tErr.Error(),
			"name":  name,
		})

		if rollbackErr := tx.Rollback(ctx); rollbackErr != nil && !errors.Is(rollbackErr, pgx.ErrTxClosed) {
			c.log.ErrorWithMetadata("unexpected error during rollback", map[string]any{
				"error": rollbackErr.Error(),
				"name":  name,
			})
			return cancel, rollbackErr
		}

		return cancel, tErr
	} else {
		if c.metrics.tm != nil {
			c.metrics.tm.IncTransactionCounter(transactionCommit, name)
		}
	}

	diff := time.Since(start)

	if c.metrics.tm != nil {
		c.metrics.tm.ObserveTransactionDurationHistogram(diff.Seconds(), name)
	}

	c.log.InfoWithMetadata("transaction success", map[string]any{
		"name": name,
	})
	return cancel, tx.Commit(ctx)
}

func (c *ConnectionPool) Query(
	ctx context.Context,
	dbFuncName string,
	sql string,
	namedArgs pgx.NamedArgs,
) (rows *pgx.Rows, cancel context.CancelFunc, err error) {
	start := time.Now()
	defer func() {
		diff := time.Since(start)
		if err != nil {
			if c.metrics.qm != nil {
				c.metrics.qm.IncQueryCounter(queryError, dbFuncName)
			}
			c.log.ErrorWithMetadata("pg query error", map[string]any{
				"error": err.Error(),
				"func":  dbFuncName,
				"sql":   sql,
			})
		} else {
			if c.metrics.qm != nil {
				c.metrics.qm.ObserveQueryDurationHistogram(diff.Seconds(), dbFuncName)
				c.metrics.qm.IncQueryCounter(querySuccess, dbFuncName)
			}
			c.log.InfoWithMetadata("pg query success", map[string]any{
				"func": dbFuncName,
				"sql":  sql,
			})
		}
	}()

	ctx, cancel = context.WithTimeout(ctx, c.queryTimeout)

	r, err := c.pool.Query(ctx, sql, namedArgs)
	if err != nil {
		if c.metrics.qm != nil {
			c.metrics.qm.IncQueryCounter(queryError, dbFuncName)
		}

		return nil, cancel, err
	}

	err = r.Err()
	if err != nil {
		return
	}

	diff := time.Since(start)
	if c.metrics.qm != nil {
		c.metrics.qm.ObserveQueryDurationHistogram(diff.Seconds(), dbFuncName)
		c.metrics.qm.IncQueryCounter(querySuccess, dbFuncName)
	}

	return &r, cancel, nil
}

func (c *ConnectionPool) QueryRow(
	ctx context.Context,
	dbFuncName string,
	sql string,
	namedArgs pgx.NamedArgs,
) (*pgx.Row, context.CancelFunc) {
	start := time.Now()

	ctx, cancel := context.WithTimeout(ctx, c.queryTimeout)

	r := c.pool.QueryRow(ctx, sql, namedArgs)

	diff := time.Since(start)
	if c.metrics.qm != nil {
		c.metrics.qm.ObserveQueryDurationHistogram(diff.Seconds(), dbFuncName)
		c.metrics.qm.IncQueryCounter(querySuccess, dbFuncName)
	}

	c.log.InfoWithMetadata("pg query operation", map[string]any{
		"func": dbFuncName,
		"sql":  sql,
	})
	return &r, cancel
}
