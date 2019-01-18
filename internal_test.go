package jobq

import (
	"testing"

	"github.com/stretchr/testify/assert"
)

func ForceWorkerN(numWorkers int) WorkerDispatcherOption {
	return func(d *WorkerDispatcher) {
		d.scaler.workersNumLowerBound = numWorkers
	}
}

func (d *WorkerDispatcher) Workers() []*Worker {
	var workers []*Worker
	for _, worker := range d.workers {
		w := (*Worker)(worker)
		workers = append(workers, w)
	}
	return workers
}

func (d *WorkerDispatcher) WorkerLen() int {
	d.workersLock.Lock()
	defer d.workersLock.Unlock()
	return len(d.workers)
}

func CheckAllWorkersStopped(t *testing.T, workers []*Worker) {
	assert := assert.New(t)

	for _, worker := range workers {
		select {
		case _, ok := <-worker.stopC:
			assert.False(ok)
		default:
			assert.Fail("worker stop incorrectly")
		}
	}
}

func CheckDispatcherStopped(t *testing.T, d *WorkerDispatcher) {
	assert := assert.New(t)

	select {
	case _, ok := <-d.stopC:
		assert.False(ok)
	default:
		assert.Fail("dispatcher stop incorrectly")
	}

	assert.Equal(0, len(d.workers))
}

func CheckJobStopped(t *testing.T, j JobTracker) {
	assert := assert.New(t)

	jb, ok := j.(*job)
	assert.True(ok)

	assert.True(jb.complete)

	select {
	case _, ok = <-jb.done:
		assert.False(ok)
	default:
		assert.Fail("job is not done")
	}

	// job context should be stopped
	select {
	case _, ok = <-jb.ctx.Done():
		assert.False(ok)
	default:
		assert.Fail("job context leaked")
	}
}
