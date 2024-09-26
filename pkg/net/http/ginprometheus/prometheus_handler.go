package ginprometheus

import (
	"github.com/gin-gonic/gin"
	"github.com/prometheus/client_golang/prometheus"
	"github.com/prometheus/client_golang/prometheus/promhttp"
)

// Handler @Summary Prometheus metrics
// @Description Expose Prometheus metrics
// @Tags Metrics
// @Produce text/plain
// @Success 200 {string} string "Prometheus metrics"
// @Router /metrics [get]
func Handler() gin.HandlerFunc {
	return HandlerFor(prometheus.DefaultGatherer)
}

func HandlerFor(g prometheus.Gatherer) gin.HandlerFunc {
	return gin.WrapH(promhttp.HandlerFor(g, promhttp.HandlerOpts{}))
}
