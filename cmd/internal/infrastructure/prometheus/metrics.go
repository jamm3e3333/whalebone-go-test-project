package prometheus

import (
	"github.com/prometheus/client_golang/prometheus"
)

type RequestMetrics struct {
	RequestCounter           *prometheus.CounterVec
	RequestDurationHistogram *prometheus.HistogramVec
}

type QueryMetrics struct {
	QueryCounter           *prometheus.CounterVec
	QueryDurationHistogram *prometheus.HistogramVec
}

type TransactionMetrics struct {
	TransactionCounter           *prometheus.CounterVec
	TransactionDurationHistogram *prometheus.HistogramVec
}

type ConnectionMetrics struct {
	DbConnectionGauge prometheus.Gauge
}

type PgMetrics struct {
	Qm *QueryMetrics
	Tm *TransactionMetrics
	Cm *ConnectionMetrics
}

func (m *RequestMetrics) IncRequestCounter(labels ...string) {
	m.RequestCounter.WithLabelValues(labels...).Inc()
}

func (m *RequestMetrics) ObserveRequestDuration(timestampDiff float64, labels ...string) {
	m.RequestDurationHistogram.WithLabelValues(labels...).Observe(timestampDiff)
}

func (m *QueryMetrics) IncQueryCounter(labels ...string) {
	m.QueryCounter.WithLabelValues(labels...).Inc()
}

func (m *TransactionMetrics) IncTransactionCounter(labels ...string) {
	m.TransactionCounter.WithLabelValues(labels...).Inc()
}

func (m *QueryMetrics) ObserveQueryDurationHistogram(timestampDiff float64, labels ...string) {
	m.QueryDurationHistogram.WithLabelValues(labels...).Observe(timestampDiff)
}

func (m *ConnectionMetrics) IncDbConnGauge() {
	m.DbConnectionGauge.Inc()
}

func (m *ConnectionMetrics) DecDbConnGauge() {
	m.DbConnectionGauge.Dec()
}

func (m *TransactionMetrics) ObserveTransactionDurationHistogram(timestampDiff float64, labels ...string) {
	m.TransactionDurationHistogram.WithLabelValues(labels...).Observe(timestampDiff)
}

type Metrics struct {
	Pm *PgMetrics
}

func NewMetrics(
	pm *PgMetrics,
) *Metrics {
	return &Metrics{
		Pm: pm,
	}
}
