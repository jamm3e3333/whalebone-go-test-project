package prometheus

import (
	"sync"

	promInfra "github.com/jamm3e3333/whalebone-go-test-project/cmd/internal/infrastructure/prometheus"
	"github.com/prometheus/client_golang/prometheus"
)

var (
	appInfo         prometheus.Labels = make(map[string]string)
	dbConn          prometheus.Labels = map[string]string{"system": "postgres"}
	uptimeTimestamp float64
)

func NewMetricsOnce(subsystem string) func() *promInfra.Metrics {
	var once sync.Once // initialization will run only once
	var metricsPtr *promInfra.Metrics = nil
	return func() *promInfra.Metrics {
		once.Do(func() {
			metricsPtr = newMetrics(subsystem)
		})

		return metricsPtr
	}
}

func newMetrics(subsystem string) *promInfra.Metrics {
	appInfoGauge := prometheus.NewGauge(prometheus.GaugeOpts{
		Subsystem:   subsystem,
		Name:        "application_info",
		ConstLabels: appInfo,
	})
	prometheus.MustRegister(appInfoGauge)
	appInfoGauge.Set(uptimeTimestamp)

	queryCounter := prometheus.NewCounterVec(
		prometheus.CounterOpts{
			Subsystem: subsystem,
			Name:      "pg_queries",
			Help:      "Number of queries executed on PG partitioned by success/error result and function name",
		},
		[]string{"result", "pg_func_name"},
	)
	prometheus.MustRegister(queryCounter)

	transactionCounter := prometheus.NewCounterVec(
		prometheus.CounterOpts{
			Subsystem: subsystem,
			Name:      "pg_transactions",
			Help:      "Number of transactions executed on PG partitioned by commit/rollback result and transaction name",
		},
		[]string{"result", "pg_transaction_name"},
	)
	prometheus.MustRegister(transactionCounter)

	queryDurationHistogram := prometheus.NewHistogramVec(
		prometheus.HistogramOpts{
			Buckets:   []float64{0.01, 0.025, 0.05, 0.075, 0.1, 0.25, 0.5, 0.75, 1, 1.5},
			Subsystem: subsystem,
			Name:      "pg_query_duration",
			Help:      "Duration of queries to PG partitioned by function name",
		},
		[]string{"pg_func_name"},
	)
	prometheus.MustRegister(queryDurationHistogram)

	dbConnectionGauge := prometheus.NewGauge(
		prometheus.GaugeOpts{
			Subsystem:   subsystem,
			Name:        "open_connections",
			ConstLabels: dbConn,
			Help:        "Count of currently open connections to postgres DB",
		})
	prometheus.MustRegister(dbConnectionGauge)
	dbConnectionGauge.Set(0)

	transactionDurationHistogram := prometheus.NewHistogramVec(
		prometheus.HistogramOpts{
			Buckets:   []float64{0.01, 0.025, 0.05, 0.075, 0.1, 0.25, 0.5, 0.75, 1, 1.5},
			Subsystem: subsystem,
			Name:      "pg_transaction_duration",
			Help:      "Duration of PG transactions partitioned by transaction name",
		}, []string{"pg_transaction_name"},
	)
	prometheus.MustRegister(transactionDurationHistogram)

	return promInfra.NewMetrics(
		&promInfra.PgMetrics{
			Qm: &promInfra.QueryMetrics{
				QueryCounter:           queryCounter,
				QueryDurationHistogram: queryDurationHistogram,
			},
			Tm: &promInfra.TransactionMetrics{
				TransactionCounter:           transactionCounter,
				TransactionDurationHistogram: transactionDurationHistogram,
			},
			Cm: &promInfra.ConnectionMetrics{DbConnectionGauge: dbConnectionGauge},
		},
	)
}
