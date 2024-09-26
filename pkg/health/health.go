package health

import (
	"sync"
	"time"

	"github.com/jamm3e3333/whalebone-go-test-project/pkg/logger"
)

type Indicator interface {
	ComponentName() string
	Status() Status
}

type Health struct {
	Indicators []Indicator
	timeout    time.Duration
	logger     logger.Logger
}

type Result struct {
	Status     Status             `json:"status" example:"up" enums:"up,down"`
	Components []*ComponentStatus `json:"components"`
}

type ComponentStatus struct {
	ComponentName string `json:"component" example:"main"`
	Status        Status `json:"status" example:"up" enums:"up,down"`
}

func NewHealthCheck(timeout time.Duration, logger logger.Logger) *Health {
	return &Health{
		Indicators: make([]Indicator, 0),
		timeout:    timeout,
		logger:     logger,
	}
}

func (h *Health) RegisterIndicator(i Indicator) {
	h.Indicators = append(h.Indicators, i)
}

func (h *Health) Handle() *Result {
	overallStatus := StatusUp
	componentsStatuses := make([]*ComponentStatus, 0)

	timeoutChan := time.After(h.timeout)

	var wg sync.WaitGroup
	var lock sync.Mutex
	var once sync.Once

	for _, i := range h.Indicators {
		wg.Add(1)
		go func(i Indicator) {
			defer func() {
				lock.Unlock()
				wg.Done()
			}()

			s := i.Status()

			if s != StatusUp {
				once.Do(func() {
					overallStatus = StatusDown
				})
			}

			lock.Lock()
			componentsStatuses = append(componentsStatuses, &ComponentStatus{
				ComponentName: i.ComponentName(),
				Status:        s,
			})
		}(i)
	}

	doneChan := make(chan interface{})
	go func(wg *sync.WaitGroup) {
		wg.Wait()
		doneChan <- struct{}{}
	}(&wg)

	select {
	case <-doneChan:
		return &Result{
			Status:     overallStatus,
			Components: componentsStatuses,
		}
	case <-timeoutChan:
		h.logger.Error("Health check had a timeout after %v!", h.timeout)

		return &Result{
			Status:     StatusTimeout,
			Components: componentsStatuses,
		}
	}
}
