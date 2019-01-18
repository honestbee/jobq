package jobq

import (
	"context"
	"sync"
	"sync/atomic"
	"time"
)

const (
	defaultJobAdjusterPeriod   time.Duration = 3 * time.Minute
	defaultMetricsReportPeriod time.Duration = 30 * time.Second
)

// JobDispatcher defines an interface for dispatching jobs
type JobDispatcher interface {
	Queue(ctx context.Context, j JobRunner) JobTracker
	QueueFunc(ctx context.Context, j JobRunnerFunc) JobTracker
	QueueTimed(ctx context.Context, j JobRunner, timeout time.Duration) JobTracker
	QueueTimedFunc(ctx context.Context, j JobRunnerFunc, timeout time.Duration) JobTracker
	Stop()
}

// NewWorkerDispatcher creates a new Dispatcher and initializes its workers.
func NewWorkerDispatcher(srvName string, opts ...WorkerDispatcherOption) *WorkerDispatcher {
	d := &WorkerDispatcher{
		workerAdjusterPeriod: defaultJobAdjusterPeriod,
		metricsReportPeriod:  defaultMetricsReportPeriod,
		stopC:                make(chan bool),
		metric:               newMetric(srvName),
		scaler:               newScaler(),
	}

	for _, opt := range opts {
		opt(d)
	}

	d.jobC = make(chan *job, d.scaler.workerPoolSize)

	d.setWorkerSize(d.scaler.workersNumLowerBound)
	d.start()

	return d
}

// WorkerDispatcher is used to maintain and delegate jobs to workers.
type WorkerDispatcher struct {
	jobC                 chan *job
	stopC                chan bool
	jobTrackerExpiry     time.Duration
	workerAdjusterPeriod time.Duration
	metricsReportPeriod  time.Duration
	metric               Metric
	scaler               *scaler
	curID                uint64

	workersLock sync.Mutex
	workers     []*worker
}

// Stop signals all workers to stop running their current
// jobs, waits for them to finish, then returns.
func (d *WorkerDispatcher) Stop() {
	d.stopC <- true
	<-d.stopC
}

// Queue takes an implementer of the JobRunner interface and schedules it to
// be run via a worker.
func (d *WorkerDispatcher) Queue(ctx context.Context, j JobRunner) JobTracker {
	ctx, cancel := context.WithCancel(ctx)
	return d.queue(ctx, cancel, j)
}

// QueueFunc is a convenience function for queuing a JobRunnerFunc
func (d *WorkerDispatcher) QueueFunc(ctx context.Context, j JobRunnerFunc) JobTracker {
	return d.Queue(ctx, JobRunner(j))
}

// QueueTimed is Queue with time limit.
func (d *WorkerDispatcher) QueueTimed(ctx context.Context, j JobRunner, timeout time.Duration) JobTracker {
	ctx, cancel := context.WithTimeout(ctx, timeout)
	return d.queue(ctx, cancel, j)
}

// QueueTimedFunc is QueueFunc with time limit.
func (d *WorkerDispatcher) QueueTimedFunc(ctx context.Context, j JobRunnerFunc, timeout time.Duration) JobTracker {
	return d.QueueTimed(ctx, JobRunner(j), timeout)
}

func (d *WorkerDispatcher) start() {
	workerAdjusterPeriod := time.NewTicker(d.workerAdjusterPeriod)
	metricsReportPeriod := time.NewTicker(d.metricsReportPeriod)
	stopWorkQueueC := make(chan struct{})

	go func() {
		for {
			select {
			case <-workerAdjusterPeriod.C:
				go d.setWorkerSize(d.scaler.scale(d.metric))
			case <-metricsReportPeriod.C:
				go d.metric.Report()
			case <-d.stopC:
				workerAdjusterPeriod.Stop()
				metricsReportPeriod.Stop()
				d.setWorkerSize(0)
				close(stopWorkQueueC)
				close(d.stopC)
				return
			}
		}
	}()
}

func (d *WorkerDispatcher) setWorkerSize(n int) {
	d.workersLock.Lock()
	defer d.workersLock.Unlock()

	lWorkers := len(d.workers)
	if lWorkers == n {
		return
	}

	for i := lWorkers; i < n; i++ {
		worker := newWorker(i, d.jobC, d.metric)
		d.workers = append(d.workers, worker)
		worker.start()
	}

	for i := n; i < lWorkers; i++ {
		d.workers[i].stop()
	}

	for i := n; i < lWorkers; i++ {
		d.workers[i].join()
		// prevent it from memory leak
		d.workers[i] = nil
	}

	d.workers = d.workers[:n]
	d.metric.SetTotalWorkers(n)
}

func (d *WorkerDispatcher) queue(ctx context.Context, cancel context.CancelFunc, j JobRunner) JobTracker {
	jobID := atomic.AddUint64(&d.curID, 1)

	job := newJob(ctx, cancel, j, uint(jobID), d.metric)
	d.jobC <- job

	d.metric.SetCurrentJobQueueSize(len(d.jobC))
	d.metric.IncJobsCounter()

	return job
}
