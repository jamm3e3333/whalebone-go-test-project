package helper

type DummyMetrics struct {
}

func NewDummyMetrics() *DummyMetrics {
	return &DummyMetrics{}
}

func (m *DummyMetrics) IncQueryCounter(_ ...string) {
}

func (m *DummyMetrics) ObserveQueryDurationHistogram(_ float64, _ ...string) {
}

func (m *DummyMetrics) IncDbConnGauge() {
}

func (m *DummyMetrics) DecDbConnGauge() {
}

func (m *DummyMetrics) IncTransactionCounter(_ ...string) {
}

func (m *DummyMetrics) ObserveTransactionDurationHistogram(_ float64, _ ...string) {
}
