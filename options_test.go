package jobq

import (
	"fmt"
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
)

func TestWorkerMargin(t *testing.T) {
	tests := []struct {
		margin      float64
		shouldPanic bool
	}{
		{0.0, true},
		{-1.0, true},
		{1.0, false},
		{1024.0, false},
	}
	for _, tt := range tests {
		tt := tt
		t.Run(fmt.Sprintf("Set worker margin:%f", tt.margin), func(t *testing.T) {
			assert := assert.New(t)

			d := &WorkerDispatcher{scaler: newScaler()}

			assert.Equal(defaultWorkerMargin, d.scaler.workerMargin)

			f := func() {
				WorkerMargin(tt.margin)(d)
			}

			if tt.shouldPanic {
				assert.Panics(f)
				assert.Equal(defaultWorkerMargin, d.scaler.workerMargin)
			} else {
				assert.NotPanics(f)
				assert.Equal(tt.margin, d.scaler.workerMargin)
			}
		})
	}
}

func TestWorkerLoadingBoundPercentage(t *testing.T) {
	tests := []struct {
		upper       float64
		lower       float64
		shouldPanic bool
	}{
		{-1000, 99, true},
		{99, -1000, true},
		{0, 99, true},
		{99, 0, true},
		{1, 99, true},
		{101, 99, true},
		{99, 101, true},
		{80, 50, false},
	}
	for _, tt := range tests {
		tt := tt
		t.Run(fmt.Sprintf("Set worker loading upper bound to %f, lower bound to %f", tt.upper, tt.lower), func(t *testing.T) {
			assert := assert.New(t)

			d := &WorkerDispatcher{scaler: newScaler()}

			assert.Equal(defaultWorkerLoadingLowerBoundPercentage, d.scaler.workerLoadingLowerBoundPercentage)
			assert.Equal(defaultWorkerLoadingUpperBoundPercentage, d.scaler.workerLoadingUpperBoundPercentage)

			f := func() {
				WorkerLoadingBoundPercentage(tt.lower, tt.upper)(d)
			}

			if tt.shouldPanic {
				assert.Panics(f)
				assert.Equal(defaultWorkerLoadingLowerBoundPercentage, d.scaler.workerLoadingLowerBoundPercentage)
				assert.Equal(defaultWorkerLoadingUpperBoundPercentage, d.scaler.workerLoadingUpperBoundPercentage)
			} else {
				assert.NotPanics(f)
				assert.Equal(tt.lower, d.scaler.workerLoadingLowerBoundPercentage)
				assert.Equal(tt.upper, d.scaler.workerLoadingUpperBoundPercentage)
			}
		})
	}
}

func TestJobTimeoutRateBoundPercentage(t *testing.T) {
	tests := []struct {
		upper       float64
		lower       float64
		shouldPanic bool
	}{
		{-1000, 99, true},
		{99, -1000, true},
		{0, 99, true},
		{99, 0, true},
		{1, 99, true},
		{101, 99, true},
		{99, 101, true},
		{80, 50, false},
	}
	for _, tt := range tests {
		tt := tt
		t.Run(fmt.Sprintf("Set job timeout rate upper bound to %f, lower bound to %f", tt.upper, tt.lower), func(t *testing.T) {
			assert := assert.New(t)

			d := &WorkerDispatcher{scaler: newScaler()}

			assert.Equal(defaultJobTimeoutRateLowerBoundPercentage, d.scaler.jobTimeoutRateLowerBoundPercentage)
			assert.Equal(defaultJobTimeoutRateUpperBoundPercentage, d.scaler.jobTimeoutRateUpperBoundPercentage)

			f := func() {
				JobTimeoutRateBoundPercentage(tt.lower, tt.upper)(d)
			}

			if tt.shouldPanic {
				assert.Panics(f)
				assert.Equal(defaultJobTimeoutRateLowerBoundPercentage, d.scaler.jobTimeoutRateLowerBoundPercentage)
				assert.Equal(defaultJobTimeoutRateUpperBoundPercentage, d.scaler.jobTimeoutRateUpperBoundPercentage)
			} else {
				assert.NotPanics(f)
				assert.Equal(tt.lower, d.scaler.jobTimeoutRateLowerBoundPercentage)
				assert.Equal(tt.upper, d.scaler.jobTimeoutRateUpperBoundPercentage)
			}
		})
	}
}

func TestMetricsReportPeriod(t *testing.T) {
	tests := []struct {
		period      time.Duration
		shouldPanic bool
	}{
		{time.Duration(-1000), true},
		{time.Duration(-1), true},
		{time.Duration(0), true},
		{time.Duration(1), false},
		{time.Duration(1000), false},
	}
	for _, tt := range tests {
		tt := tt
		t.Run(fmt.Sprintf("Set metrics report period to %s", tt.period), func(t *testing.T) {
			assert := assert.New(t)

			d := &WorkerDispatcher{}

			assert.Equal(time.Duration(0), d.metricsReportPeriod)

			f := func() {
				MetricsReportPeriod(tt.period)(d)
			}

			if tt.shouldPanic {
				assert.Panics(f)
				assert.Equal(time.Duration(0), d.metricsReportPeriod)
			} else {
				assert.NotPanics(f)
				assert.Equal(tt.period, d.metricsReportPeriod)
			}
		})
	}
}

func TestWorkerAdjustPeriod(t *testing.T) {
	tests := []struct {
		period      time.Duration
		shouldPanic bool
	}{
		{time.Duration(-1000), true},
		{time.Duration(-1), true},
		{time.Duration(0), true},
		{time.Duration(1), false},
		{time.Duration(1000), false},
	}
	for _, tt := range tests {
		tt := tt
		t.Run(fmt.Sprintf("Set worker adjust period to %s", tt.period), func(t *testing.T) {
			assert := assert.New(t)

			d := &WorkerDispatcher{}

			assert.Equal(time.Duration(0), d.workerAdjusterPeriod)

			f := func() {
				WorkerAdjustPeriod(tt.period)(d)
			}

			if tt.shouldPanic {
				assert.Panics(f)
				assert.Equal(time.Duration(0), d.workerAdjusterPeriod)
			} else {
				assert.NotPanics(f)
				assert.Equal(tt.period, d.workerAdjusterPeriod)
			}
		})
	}
}

func TestWorkerN(t *testing.T) {
	tests := []struct {
		worker      int
		shouldPanic bool
	}{
		{0, true},
		{-1, true},
		{1, false},
		{1024, false},
	}
	for _, tt := range tests {
		tt := tt
		t.Run(fmt.Sprintf("Set %d worker(s)", tt.worker), func(t *testing.T) {
			assert := assert.New(t)

			d := &WorkerDispatcher{scaler: newScaler()}

			assert.Equal(defaultWorkersNumLowerBound, d.scaler.workersNumLowerBound)

			f := func() {
				WorkerN(tt.worker)(d)
			}

			if tt.shouldPanic {
				assert.Panics(f)
				assert.Equal(defaultWorkersNumLowerBound, d.scaler.workersNumLowerBound)
			} else {
				assert.NotPanics(f)
				assert.Equal(tt.worker, d.scaler.workersNumLowerBound)
			}
		})
	}
}

func TestWorkerPoolSize(t *testing.T) {
	tests := []struct {
		size        int
		shouldPanic bool
	}{
		{0, true},
		{-1, true},
		{1, false},
		{1024, false},
	}
	for _, tt := range tests {
		tt := tt
		t.Run(fmt.Sprintf("Set pool size: %d ", tt.size), func(t *testing.T) {
			assert := assert.New(t)

			d := &WorkerDispatcher{scaler: newScaler()}

			assert.Equal(defaultWorkerPoolSize, d.scaler.workerPoolSize)

			f := func() {
				WorkerPoolSize(tt.size)(d)
			}

			if tt.shouldPanic {
				assert.Panics(f)
				assert.Equal(defaultWorkerPoolSize, d.scaler.workerPoolSize)
			} else {
				assert.NotPanics(f)
				assert.Equal(tt.size, d.scaler.workerPoolSize)
			}
		})
	}
}

func TestJobStatusExpiry(t *testing.T) {
	tests := []struct {
		expiry      time.Duration
		shouldPanic bool
	}{
		{time.Duration(-1000), true},
		{time.Duration(-1), true},
		{time.Duration(0), true},
		{time.Duration(1), false},
		{time.Duration(1000), false},
	}
	for _, tt := range tests {
		tt := tt
		t.Run(fmt.Sprintf("Set job status expiry to %s", tt.expiry), func(t *testing.T) {
			assert := assert.New(t)

			d := &WorkerDispatcher{}

			assert.Equal(time.Duration(0), d.jobTrackerExpiry)

			f := func() {
				JobStatusExpiry(tt.expiry)(d)
			}

			if tt.shouldPanic {
				assert.Panics(f)
				assert.Equal(time.Duration(0), d.jobTrackerExpiry)
			} else {
				assert.NotPanics(f)
				assert.Equal(tt.expiry, d.jobTrackerExpiry)
			}
		})
	}
}
