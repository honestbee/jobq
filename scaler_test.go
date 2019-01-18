package jobq

import (
	"testing"
)

type mockMetrics struct {
	workerLoading  float64
	jobTimeoutRate float64
	totalWorkers   uint64
}

func (m *mockMetrics) SetTotalWorkers(v int)        {}
func (m *mockMetrics) TotalWorkers() uint64         { return m.totalWorkers }
func (m *mockMetrics) SetCurrentJobQueueSize(v int) {}
func (m *mockMetrics) IncBusyWorker()               {}
func (m *mockMetrics) DecBusyWorker()               {}
func (m *mockMetrics) IncJobTimeout()               {}
func (m *mockMetrics) IncJobsCounter()              {}
func (m *mockMetrics) Report()                      {}
func (m *mockMetrics) JobTimeoutRate() float64      { return m.jobTimeoutRate }
func (m *mockMetrics) WorkerLoading() float64       { return m.workerLoading }
func (m *mockMetrics) ResetCounters()               {}

func TestScale(t *testing.T) {
	tests := []struct {
		description                        string
		jobTimeoutRateUpperBoundPercentage float64
		jobTimeoutRateLowerBoundPercentage float64
		workerLoadingUpperBoundPercentage  float64
		workerLoadingLowerBoundPercentage  float64
		workersNumUpperBound               int
		workersNumLowerBound               int
		workerMargin                       float64
		metrics                            *mockMetrics
		want                               int
	}{
		{
			description:                        "testing worker loading over upper bound",
			jobTimeoutRateUpperBoundPercentage: 90,
			jobTimeoutRateLowerBoundPercentage: 80,
			workerLoadingUpperBoundPercentage:  5,
			workerLoadingLowerBoundPercentage:  1,
			workersNumUpperBound:               30,
			workersNumLowerBound:               10,
			workerMargin:                       2,
			metrics: &mockMetrics{
				totalWorkers:   10,
				workerLoading:  0.1,
				jobTimeoutRate: 0.0,
			},
			want: 20,
		},
		{
			description:                        "testing job timeouts over upper bound",
			jobTimeoutRateUpperBoundPercentage: 5,
			jobTimeoutRateLowerBoundPercentage: 1,
			workerLoadingUpperBoundPercentage:  90,
			workerLoadingLowerBoundPercentage:  80,
			workersNumUpperBound:               30,
			workersNumLowerBound:               10,
			workerMargin:                       2,
			metrics: &mockMetrics{
				totalWorkers:   10,
				workerLoading:  0.7,
				jobTimeoutRate: 0.2,
			},
			want: 20,
		},
		{
			description:                        "testing job timeout and worker loading in balance",
			jobTimeoutRateUpperBoundPercentage: defaultJobTimeoutRateUpperBoundPercentage,
			jobTimeoutRateLowerBoundPercentage: defaultJobTimeoutRateLowerBoundPercentage,
			workerLoadingUpperBoundPercentage:  defaultWorkerLoadingUpperBoundPercentage,
			workerLoadingLowerBoundPercentage:  defaultWorkerLoadingLowerBoundPercentage,
			workersNumUpperBound:               30,
			workersNumLowerBound:               10,
			workerMargin:                       2,
			metrics: &mockMetrics{
				totalWorkers:   10,
				workerLoading:  0.7,
				jobTimeoutRate: 0.6,
			},
			want: 10,
		},
		{
			description:                        "testing worker decreasing",
			jobTimeoutRateUpperBoundPercentage: defaultJobTimeoutRateUpperBoundPercentage,
			jobTimeoutRateLowerBoundPercentage: defaultJobTimeoutRateLowerBoundPercentage,
			workerLoadingUpperBoundPercentage:  defaultWorkerLoadingUpperBoundPercentage,
			workerLoadingLowerBoundPercentage:  defaultWorkerLoadingLowerBoundPercentage,
			workersNumUpperBound:               30,
			workersNumLowerBound:               10,
			workerMargin:                       2,
			metrics: &mockMetrics{
				totalWorkers:   20,
				workerLoading:  0.1,
				jobTimeoutRate: 0.1,
			},
			want: 10,
		},
		{
			description:                        "testing worker increasing over worker num upper bound",
			jobTimeoutRateUpperBoundPercentage: defaultJobTimeoutRateUpperBoundPercentage,
			jobTimeoutRateLowerBoundPercentage: defaultJobTimeoutRateLowerBoundPercentage,
			workerLoadingUpperBoundPercentage:  defaultWorkerLoadingUpperBoundPercentage,
			workerLoadingLowerBoundPercentage:  defaultWorkerLoadingLowerBoundPercentage,
			workersNumUpperBound:               30,
			workersNumLowerBound:               10,
			workerMargin:                       10,
			metrics: &mockMetrics{
				totalWorkers:   20,
				workerLoading:  0.9,
				jobTimeoutRate: 0.1,
			},
			want: 30,
		},
		{
			description:                        "testing worker decreasing over worker num lower bound",
			jobTimeoutRateUpperBoundPercentage: defaultJobTimeoutRateUpperBoundPercentage,
			jobTimeoutRateLowerBoundPercentage: defaultJobTimeoutRateLowerBoundPercentage,
			workerLoadingUpperBoundPercentage:  defaultWorkerLoadingUpperBoundPercentage,
			workerLoadingLowerBoundPercentage:  defaultWorkerLoadingLowerBoundPercentage,
			workersNumUpperBound:               30,
			workersNumLowerBound:               10,
			workerMargin:                       10,
			metrics: &mockMetrics{
				totalWorkers:   20,
				workerLoading:  0.1,
				jobTimeoutRate: 0.1,
			},
			want: 10,
		},
	}

	for _, tt := range tests {
		tt := tt
		t.Run(tt.description, func(t *testing.T) {
			sc := newScaler()
			sc.jobTimeoutRateUpperBoundPercentage = tt.jobTimeoutRateUpperBoundPercentage
			sc.jobTimeoutRateLowerBoundPercentage = tt.jobTimeoutRateLowerBoundPercentage
			sc.workerLoadingUpperBoundPercentage = tt.workerLoadingUpperBoundPercentage
			sc.workerLoadingLowerBoundPercentage = tt.workerLoadingLowerBoundPercentage
			sc.workersNumUpperBound = tt.workersNumUpperBound
			sc.workersNumLowerBound = tt.workersNumLowerBound
			sc.workerMargin = tt.workerMargin

			got := sc.scale(tt.metrics)
			if tt.want != got {
				t.Errorf("want:%d, got:%d", tt.want, got)
			}
		})
	}
}
