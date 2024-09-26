package health

import (
	"net/http"

	"github.com/gin-gonic/gin"
	healthcheck "github.com/jamm3e3333/whalebone-go-test-project/pkg/health"
)

type CheckHandler interface {
	Handle() *healthcheck.Result
}

type Controller struct {
	readinessService CheckHandler
	livenessService  CheckHandler
	engine           *gin.Engine
}

func NewController(rs CheckHandler, ls CheckHandler, e *gin.Engine) *Controller {
	return &Controller{rs, ls, e}
}

func (c *Controller) Register(ctx *gin.Engine) {
	ctx.GET("health/readiness", c.HandleHealthCheckReadiness)
	ctx.GET("health/liveness", c.HandleHealthCheckLiveness)
}

// HandleHealthCheckReadiness @Summary Health check for readiness probe
// @Description Health check of the application
// @Tags Health
// @Produce json
// @Success 200 {object} healthcheck.Result
// @Success 503 {object} healthcheck.Result
// @Router /health/readiness [get]
func (c *Controller) HandleHealthCheckReadiness(ctx *gin.Context) {
	code := http.StatusOK

	result := c.readinessService.Handle()

	if result.Status != healthcheck.StatusUp {
		code = http.StatusServiceUnavailable
	}

	ctx.JSON(code, result)
}

// HandleHealthCheckLiveness @Summary Health check for liveness probe
// @Description Health check for liveness probe
// @Tags Health
// @Produce json
// @Success 200 {object} healthcheck.Result
// @Success 503 {object} healthcheck.Result
// @Router /health/liveness [get]
func (c *Controller) HandleHealthCheckLiveness(ctx *gin.Context) {
	code := http.StatusOK

	result := c.livenessService.Handle()

	if result.Status != healthcheck.StatusUp {
		code = http.StatusServiceUnavailable
	}

	ctx.JSON(code, result)
}
