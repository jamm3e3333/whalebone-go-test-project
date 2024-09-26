package pgx

type MonitoringMetrics struct {
	qm QueryMetrics
	tm TransactionMetrics
}

type QueryMetrics interface {
	ObserveQueryDurationHistogram(timestampDiff float64, labels ...string)
	IncQueryCounter(labels ...string)
}

type TransactionMetrics interface {
	ObserveTransactionDurationHistogram(timestampDiff float64, labels ...string)
	IncTransactionCounter(labels ...string)
}

type ConnectionMetrics interface {
	IncDbConnGauge()
	DecDbConnGauge()
}
