package jobq

import (
	"fmt"
	"math"
	"testing"
)

func TestGetTotalWorkers(t *testing.T) {
	tests := []struct {
		set  int
		want uint64
	}{
		{
			set:  10,
			want: 10,
		},
		{
			set:  -1,
			want: 0,
		},
	}

	for _, tt := range tests {
		tt := tt
		t.Run(fmt.Sprintf("set:%d", tt.set), func(t *testing.T) {
			metrics := newMetric(func(TrackParams) {})
			metrics.SetTotalWorkers(tt.set)
			got := metrics.TotalWorkers()
			if tt.want != got {
				t.Errorf("want:%d, got:%d", tt.want, got)
			}
		})
	}
}

func TestJobTimeoutRate(t *testing.T) {
	tests := []struct {
		description string
		presets     []func(m Metric)
		want        float64
	}{
		{
			description: "testing NaN case",
			want:        math.NaN(),
		},
		{
			description: "testing normal case",
			presets: []func(m Metric){
				func(m Metric) { m.IncJobTimeout() },
				func(m Metric) { m.IncJobTimeout() },
				func(m Metric) { m.IncJobsCounter() },
				func(m Metric) { m.IncJobsCounter() },
				func(m Metric) { m.IncJobsCounter() },
			},
			want: float64(2) / float64(3),
		},
	}

	for _, tt := range tests {
		tt := tt
		t.Run(tt.description, func(t *testing.T) {
			metrics := newMetric(func(TrackParams) {})
			for _, f := range tt.presets {
				f(metrics)
			}

			got := metrics.JobTimeoutRate()
			if math.IsNaN(tt.want) {
				if !math.IsNaN(got) {
					t.Errorf("want:NaN, got:%v", got)
				}
			} else if tt.want != got {
				t.Errorf("want:%f, got:%f", tt.want, got)
			}
		})
	}
}

func TestWorkerLoading(t *testing.T) {
	tests := []struct {
		description string
		presets     []func(m Metric)
		want        float64
	}{
		{
			description: "testing NaN case",
			want:        math.NaN(),
		},
		{
			description: "testing normal case",
			presets: []func(m Metric){
				func(m Metric) { m.IncBusyWorker() },
				func(m Metric) { m.IncBusyWorker() },
				func(m Metric) { m.IncBusyWorker() },
				func(m Metric) { m.DecBusyWorker() },
				func(m Metric) { m.SetTotalWorkers(3) },
			},
			want: float64(2) / float64(3),
		},
	}

	for _, tt := range tests {
		tt := tt
		t.Run(tt.description, func(t *testing.T) {
			metrics := newMetric(func(TrackParams) {})
			for _, f := range tt.presets {
				f(metrics)
			}

			got := metrics.WorkerLoading()
			if math.IsNaN(tt.want) {
				if !math.IsNaN(got) {
					t.Errorf("want:NaN, got:%v", got)
				}
			} else if tt.want != got {
				t.Errorf("want:%f, got:%f", tt.want, got)
			}
		})
	}
}

func TestResetCounters(t *testing.T) {
	tests := []struct {
		presets []func(m Metric)
	}{
		{
			presets: []func(m Metric){
				func(m Metric) { m.IncJobTimeout() },
				func(m Metric) { m.IncJobTimeout() },
				func(m Metric) { m.IncJobsCounter() },
				func(m Metric) { m.IncJobsCounter() },
				func(m Metric) { m.IncJobsCounter() },
			},
		},
	}

	for i, tt := range tests {
		tt := tt
		t.Run(fmt.Sprintf("case[%d]", i), func(t *testing.T) {
			metrics := newMetric(func(TrackParams) {})
			for _, f := range tt.presets {
				f(metrics)
			}

			if math.IsNaN(metrics.JobTimeoutRate()) {
				t.Errorf("before reset counters want job timeout rate != 0")
			} else {
				metrics.ResetCounters()
				if !math.IsNaN(metrics.JobTimeoutRate()) {
					t.Errorf("after reset counters want job timeout rate == 0")
				}
			}
		})
	}
}
