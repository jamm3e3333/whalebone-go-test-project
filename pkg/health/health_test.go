package health

import (
	"fmt"

	"github.com/jamm3e3333/whalebone-go-test-project/pkg/logger"
	"github.com/stretchr/testify/assert"

	"testing"
	"time"
)

type HealthIndicatorMock struct {
	Name       string
	StatusFunc func() Status
}

func (i *HealthIndicatorMock) ComponentName() string {
	return i.Name
}

func (i *HealthIndicatorMock) Status() Status {
	return i.StatusFunc()
}

func Test_HealthCheck_Handle_Timeout(t *testing.T) {
	lg := logger.New(logger.ParseLevel("debug"), false)

	timeout := 50 * time.Millisecond
	hc := NewHealthCheck(timeout, lg)

	hi := HealthIndicatorMock{
		Name: "test",
		StatusFunc: func() Status {
			time.Sleep(2 * time.Second)
			return StatusUp
		},
	}
	hc.RegisterIndicator(&hi)

	hcr := hc.Handle()

	assert.Equal(t, StatusTimeout, hcr.Status)
}

func Test_HealthCheck_Handle_Up(t *testing.T) {
	lg := logger.New(logger.ParseLevel("debug"), false)

	timeout := 500 * time.Millisecond
	hc := NewHealthCheck(timeout, lg)

	hi := HealthIndicatorMock{
		Name: "test",
		StatusFunc: func() Status {
			return StatusUp
		},
	}
	hc.RegisterIndicator(&hi)

	hcr := hc.Handle()

	expectedComponentsResult := []*ComponentStatus{{
		ComponentName: "test",
		Status:        StatusUp,
	}}
	assert.Equal(t, expectedComponentsResult, hcr.Components)

	assert.Equal(t, StatusUp, hcr.Status)
}

func Test_HealthCheck_Handle_Down(t *testing.T) {
	lg := logger.New(logger.ParseLevel("debug"), false)

	timeout := 500 * time.Millisecond
	hc := NewHealthCheck(timeout, lg)

	hi := HealthIndicatorMock{
		Name: "test",
		StatusFunc: func() Status {
			return StatusDown
		},
	}
	hc.RegisterIndicator(&hi)

	hcr := hc.Handle()

	expectedComponentsResult := []*ComponentStatus{{
		ComponentName: "test",
		Status:        StatusDown,
	}}
	assert.Equal(t, expectedComponentsResult, hcr.Components)

	assert.Equal(t, StatusDown, hcr.Status)
}

func Test_HealthCheck_With_Multiple_Indicators_Handle_Up(t *testing.T) {
	lg := logger.New(logger.ParseLevel("debug"), false)

	timeout := 500 * time.Millisecond
	hc := NewHealthCheck(timeout, lg)

	hi1 := HealthIndicatorMock{
		Name: "test1",
		StatusFunc: func() Status {
			return StatusUp
		},
	}
	hc.RegisterIndicator(&hi1)

	hi2 := HealthIndicatorMock{
		Name: "test2",
		StatusFunc: func() Status {
			return StatusUp
		},
	}
	hc.RegisterIndicator(&hi2)

	hi3 := HealthIndicatorMock{
		Name: "test3",
		StatusFunc: func() Status {
			return StatusUp
		},
	}
	hc.RegisterIndicator(&hi3)

	hcr := hc.Handle()

	expectedComponentsResults := map[string]Status{
		"test1": StatusUp,
		"test2": StatusUp,
		"test3": StatusUp,
	}
	actualResults := make(map[string]Status)

	for _, componentStatus := range hcr.Components {
		actualResults[componentStatus.ComponentName] = componentStatus.Status
	}

	for componentName, expectedStatus := range expectedComponentsResults {
		actualStatus, ok := actualResults[componentName]
		assert.True(t, ok, fmt.Sprintf("Missing expected result for component: %s", componentName))
		assert.Equal(t, expectedStatus, actualStatus)
	}

	assert.Equal(t, StatusUp, hcr.Status)
}

func Test_HealthCheck_With_Multiple_Indicators_Handle_Down(t *testing.T) {
	lg := logger.New(logger.ParseLevel("debug"), false)

	timeout := 500 * time.Millisecond
	hc := NewHealthCheck(timeout, lg)

	hi1 := HealthIndicatorMock{
		Name: "test1",
		StatusFunc: func() Status {
			return StatusUp
		},
	}
	hc.RegisterIndicator(&hi1)

	hi2 := HealthIndicatorMock{
		Name: "test2",
		StatusFunc: func() Status {
			return StatusDown
		},
	}
	hc.RegisterIndicator(&hi2)

	hi3 := HealthIndicatorMock{
		Name: "test3",
		StatusFunc: func() Status {
			return StatusUp
		},
	}
	hc.RegisterIndicator(&hi3)

	hcr := hc.Handle()

	expectedComponentsResults := map[string]Status{
		"test1": StatusUp,
		"test2": StatusDown,
		"test3": StatusUp,
	}
	actualResults := make(map[string]Status)

	for _, componentStatus := range hcr.Components {
		actualResults[componentStatus.ComponentName] = componentStatus.Status
	}

	for componentName, expectedStatus := range expectedComponentsResults {
		actualStatus, ok := actualResults[componentName]
		assert.True(t, ok, fmt.Sprintf("Missing expected result for component: %s", componentName))
		assert.Equal(t, expectedStatus, actualStatus)
	}

	assert.Equal(t, StatusDown, hcr.Status)
}
