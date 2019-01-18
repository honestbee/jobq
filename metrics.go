package jobq

import (
	"sync/atomic"

	"github.com/honestbee/orochi/internal/prometheus"
)

// Metric represents the contract that it must report corresponding metrics.
type Metric interface {
	SetTotalWorkers(v int)
	TotalWorkers() uint64
	SetCurrentJobQueueSize(v int)
	IncBusyWorker()
	DecBusyWorker()
	IncJobTimeout()
	IncJobsCounter()
	Report()
	JobTimeoutRate() float64
	WorkerLoading() float64
	ResetCounters()
}

type metric struct {
	collector prometheus.Collector

	// for prometheus metrics usage
	curJobQueueSize uint64
	totalWorkers    uint64
	busyWorkers     uint64

	// for calculating adjust workers
	jobTimeouts uint64
	jobsCounter uint64
}

func newMetric(name string) Metric {
	return &metric{
		collector: prometheus.Metrics.Get(name),
	}
}

func (m *metric) Report() {
	m.collector.Get(prometheus.WorkerpoolCurrentJobQueueSize).Set(float64(atomic.LoadUint64(&m.curJobQueueSize)))
	m.collector.Get(prometheus.WorkerpoolTotalWorkers).Set(float64(atomic.LoadUint64(&m.totalWorkers)))
	m.collector.Get(prometheus.WorkerpoolBusyWorkers).Set(float64(atomic.LoadUint64(&m.busyWorkers)))
}

func (m *metric) TotalWorkers() uint64 {
	return atomic.LoadUint64(&m.totalWorkers)
}

func (m *metric) JobTimeoutRate() float64 {
	return (float64)(atomic.LoadUint64(&m.jobTimeouts)) / (float64)(atomic.LoadUint64(&m.jobsCounter))
}

func (m *metric) WorkerLoading() float64 {
	return (float64)(atomic.LoadUint64(&m.busyWorkers)) / (float64)(atomic.LoadUint64(&m.totalWorkers))
}

func (m *metric) ResetCounters() {
	atomic.StoreUint64(&m.jobTimeouts, 0)
	atomic.StoreUint64(&m.jobsCounter, 0)
}

func (m *metric) IncJobsCounter() {
	atomic.AddUint64(&m.jobsCounter, 1)
}

func (m *metric) IncJobTimeout() {
	atomic.AddUint64(&m.jobTimeouts, 1)
}

func (m *metric) SetTotalWorkers(v int) {
	if v >= 0 {
		atomic.StoreUint64(&m.totalWorkers, uint64(v))
	}
}

func (m *metric) SetCurrentJobQueueSize(v int) {
	if v >= 0 {
		atomic.StoreUint64(&m.curJobQueueSize, uint64(v))
	}
}

func (m *metric) IncBusyWorker() {
	atomic.AddUint64(&m.busyWorkers, 1)
}

func (m *metric) DecBusyWorker() {
	atomic.AddUint64(&m.busyWorkers, ^uint64(0))
}
