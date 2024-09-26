package ginprometheus

import (
	"io"
	"net/http"
	"strconv"
	"time"

	"github.com/gin-gonic/gin"
	"github.com/prometheus/client_golang/prometheus"
)

const (
	LabelURL        = "url"
	LabelMethod     = "method"
	LabelStatusCode = "code"
)

var defaultLabels = []Label{
	{LabelURL, getURLForLabel},
	{LabelMethod, getMethodForLabel},
	{LabelStatusCode, getStatusForLabel},
}

var defaultRequestDurationBuckets = []float64{0.05, 0.1, 0.2, 0.4, 0.6, 1.0, 1.5, 2.0}

func getStatusForLabel(c *gin.Context) string {
	return strconv.Itoa(c.Writer.Status())
}

func getURLForLabel(c *gin.Context) string {
	return c.FullPath()
}

func getMethodForLabel(c *gin.Context) string {
	// we filter only standard methods to prevent
	// metrics cardinality explosion
	switch c.Request.Method {
	case http.MethodGet:
	case http.MethodHead:
	case http.MethodPost:
	case http.MethodPut:
	case http.MethodPatch:
	case http.MethodDelete:
	case http.MethodConnect:
	case http.MethodOptions:
	case http.MethodTrace:
	default:
		return "other"
	}

	return c.Request.Method
}

type CustomLabelFunc func(c *gin.Context) string

type Label struct {
	Name string
	Func CustomLabelFunc
}

type Config struct {
	Namespace string
	Subsystem string
	Labels    []Label

	RequestDurationHistogramBuckets []float64
	RequestSizeSummaryObjectvies    map[float64]float64
	ResponseSizeSummaryObjectvies   map[float64]float64
}

type metricsHandler struct {
	Config

	labelFuncs []CustomLabelFunc

	reqCount           *prometheus.CounterVec
	reqDurationSeconds *prometheus.HistogramVec
	reqSize            *prometheus.SummaryVec
	respSize           *prometheus.SummaryVec
}

func buildLabels(labels ...Label) ([]string, []CustomLabelFunc) {
	labelPosition := make(map[string]int)
	labelNames := make([]string, 0, len(labels))
	labelFuncs := make([]CustomLabelFunc, 0, len(labels))

	for _, l := range labels {
		if pos, ok := labelPosition[l.Name]; ok {
			labelFuncs[pos] = l.Func
			continue
		}

		labelNames = append(labelNames, l.Name)
		labelFuncs = append(labelFuncs, l.Func)
		labelPosition[l.Name] = len(labelFuncs) - 1
	}

	return labelNames, labelFuncs
}

func concatLabels(labelSlices ...[]Label) []Label {
	size := 0
	for _, l := range labelSlices {
		size += len(l)
	}

	labels := make([]Label, 0, size)

	for _, l := range labelSlices {
		labels = append(labels, l...)
	}

	return labels
}

func Measure(config Config) gin.HandlerFunc {
	return MeasureWith(prometheus.DefaultRegisterer, config)
}

func MeasureWith(r prometheus.Registerer, config Config) gin.HandlerFunc {
	m := &metricsHandler{
		Config: config,
	}
	labels := concatLabels(defaultLabels, m.Labels)
	labelNames, labelFuncs := buildLabels(labels...)
	m.labelFuncs = labelFuncs

	m.reqCount = prometheus.NewCounterVec(
		prometheus.CounterOpts{
			Namespace: m.Namespace,
			Subsystem: m.Subsystem,
			Name:      "requests_total",
			Help:      "Total number of requests",
		},
		labelNames,
	)
	reqDurationBuckets := m.RequestDurationHistogramBuckets
	if reqDurationBuckets == nil {
		reqDurationBuckets = defaultRequestDurationBuckets
	}
	m.reqDurationSeconds = prometheus.NewHistogramVec(
		prometheus.HistogramOpts{
			Namespace: m.Namespace,
			Subsystem: m.Subsystem,
			Name:      "request_duration_seconds",
			Help:      "Duration of HTTP handler exection",

			Buckets: reqDurationBuckets,
		},
		labelNames,
	)
	m.reqSize = prometheus.NewSummaryVec(
		prometheus.SummaryOpts{
			Namespace: m.Namespace,
			Subsystem: m.Subsystem,
			Name:      "request_size_bytes",
			Help:      "Estimated number of bytes processed by HTTP handler",

			Objectives: m.RequestSizeSummaryObjectvies,
		},
		labelNames,
	)
	m.respSize = prometheus.NewSummaryVec(
		prometheus.SummaryOpts{
			Namespace: m.Namespace,
			Subsystem: m.Subsystem,
			Name:      "response_size_bytes",
			Help:      "Size of HTTP response body returned to the client",

			Objectives: m.ResponseSizeSummaryObjectvies,
		},
		labelNames,
	)

	metrics := []prometheus.Collector{
		m.reqCount,
		m.reqDurationSeconds,
		m.reqSize,
		m.respSize,
	}
	for _, m := range metrics {
		r.MustRegister(m)
	}

	return m.record
}

func (m *metricsHandler) record(c *gin.Context) {
	requestStart := time.Now()

	requestBodySize, ok := getRequestStaticBodySize(c.Request)
	var requestBodySizeCounter *countingReader
	if !ok {
		requestBodySizeCounter = &countingReader{
			r: c.Request.Body,
		}
		c.Request.Body = requestBodySizeCounter
	}

	c.Next()

	if requestBodySizeCounter != nil {
		requestBodySize = requestBodySizeCounter.Count
	}

	requestSize := estimateRequestSize(c.Request, requestBodySize)
	requestDurationSeconds := time.Since(requestStart).Seconds()

	responseSize := c.Writer.Size()
	if responseSize < 0 {
		responseSize = 0
	}

	labels := make([]string, len(m.labelFuncs))
	for i, labelFn := range m.labelFuncs {
		labels[i] = labelFn(c)
	}

	m.reqDurationSeconds.WithLabelValues(labels...).Observe(requestDurationSeconds)
	m.reqCount.WithLabelValues(labels...).Inc()
	m.reqSize.WithLabelValues(labels...).Observe(float64(requestSize))
	m.respSize.WithLabelValues(labels...).Observe(float64(responseSize))
}

func getRequestStaticBodySize(req *http.Request) (n int64, ok bool) {
	if req.ContentLength >= 0 {
		return req.ContentLength, true
	}
	return -1, false
}

func estimateRequestSize(r *http.Request, bodySize int64) int64 {
	s := 0

	s += len(r.Method)

	s += len(r.Proto)

	for name, values := range r.Header {
		s += len(name)
		for _, v := range values {
			s += len(v)
		}
	}

	s += len(r.URL.RawPath)
	s += len(r.URL.RawQuery)

	return int64(s) + bodySize
}

type countingReader struct {
	r     io.Reader
	Count int64
}

func (cr *countingReader) Read(buf []byte) (n int, err error) {
	n, err = cr.r.Read(buf)
	cr.Count += int64(n)
	return
}

func (cr *countingReader) Close() error {
	switch r := cr.r.(type) {
	case io.ReadCloser:
		return r.Close()
	}
	return nil
}
