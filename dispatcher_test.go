package jobq_test

import (
	"context"
	"fmt"
	"testing"
	"time"

	"github.com/pkg/errors"
	"github.com/stretchr/testify/assert"

	"github.com/honestbee/orochi/internal/jobq"
	"github.com/honestbee/orochi/internal/prometheus"
)

func init() {
	prometheus.Init()
}

func TestWorkerDispatcher_Queue(t *testing.T) {
	qf := func(d jobq.JobDispatcher, f jobq.JobRunnerFunc) jobq.JobTracker {
		return d.Queue(context.Background(), jobq.JobRunner(f))
	}
	testQueueBasic(t, qf)
	testQueueCancel(t, qf)
	testQueueStop(t, qf)
}
func TestWorkerDispatcher_QueueFunc(t *testing.T) {
	qf := func(d jobq.JobDispatcher, f jobq.JobRunnerFunc) jobq.JobTracker {
		return d.QueueFunc(context.Background(), f)
	}
	testQueueBasic(t, qf)
	testQueueCancel(t, qf)
	testQueueStop(t, qf)
}

func TestWorkerDispatcher_QueueTimed(t *testing.T) {
	qf := func(d jobq.JobDispatcher, f jobq.JobRunnerFunc, timeout time.Duration) jobq.JobTracker {
		return d.QueueTimed(context.Background(), jobq.JobRunner(f), timeout)
	}
	testQueueTimeout(t, qf)
	testQueueTimeout_completeBeforeTimeout(t, qf)
}

func TestWorkerDispatcher_QueueTimedFunc(t *testing.T) {
	qf := func(d jobq.JobDispatcher, f jobq.JobRunnerFunc, timeout time.Duration) jobq.JobTracker {
		return d.QueueTimedFunc(context.Background(), f, timeout)
	}
	testQueueTimeout(t, qf)
	testQueueTimeout_completeBeforeTimeout(t, qf)
}

func TestWorkerDispatcher_EarlyStop(t *testing.T) {
	tests := []struct {
		workerN int
	}{
		{1},
		{2},
		{4},
		{8},
	}
	for _, tt := range tests {
		tt := tt
		t.Run(fmt.Sprintf("With %d worker(s)", tt.workerN), func(t *testing.T) {
			assert := assert.New(t)

			d := jobq.NewWorkerDispatcher("orochi_workers", jobq.WorkerN(tt.workerN))
			workers := d.Workers()
			for i, worker := range workers {
				assert.Equal(i, worker.ID())
			}
			d.Stop()

			jobq.CheckDispatcherStopped(t, d)
			assert.Equal(tt.workerN, len(workers))
			assert.Equal(0, d.WorkerLen())

			jobq.CheckAllWorkersStopped(t, workers)
		})
	}
}

type queueFunc func(d jobq.JobDispatcher, f jobq.JobRunnerFunc) jobq.JobTracker
type queueTimedFunc func(d jobq.JobDispatcher, f jobq.JobRunnerFunc, timeout time.Duration) jobq.JobTracker

func testQueueBasic(t *testing.T, qf queueFunc) {
	tests := []struct {
		payload      interface{}
		err          error
		want         interface{}
		errAssertion assert.ErrorAssertionFunc
	}{
		{"foo", nil, "foo", assert.NoError},
		{nil, nil, nil, assert.NoError},
		{nil, errors.New("empty"), nil, assert.Error},
	}
	for _, tt := range tests {
		tt := tt
		t.Run(fmt.Sprintf("Queue a job runner with %v, %v returned", tt.payload, tt.err), func(t *testing.T) {
			d := jobq.NewWorkerDispatcher("orochi_workers")
			defer d.Stop()

			proceed := make(chan bool, 1)

			f := func(ctx context.Context) (interface{}, error) {
				<-proceed
				return tt.payload, tt.err
			}

			assert := assert.New(t)

			tracker := qf(d, f)

			status := tracker.Status()
			assert.False(status.Complete)
			assert.False(status.Success)
			assert.Equal("", status.Error)
			assert.Nil(status.Payload)

			proceed <- true
			payload, err := tracker.Result()
			tt.errAssertion(t, err)
			assert.Equal(tt.want, payload)

			jobq.CheckJobStopped(t, tracker)

			status = tracker.Status()
			assert.True(status.Complete)
			if tt.err == nil {
				assert.True(status.Success)
				assert.Equal("", status.Error)
			} else {
				assert.False(status.Success)
				assert.Equal(tt.err.Error(), status.Error)
			}
			assert.Equal(tt.want, status.Payload)
			assert.True(status.FinishedAt.After(status.StartedAt))
		})
	}
}

func testQueueCancel(t *testing.T, qf queueFunc) {
	assert := assert.New(t)

	d := jobq.NewWorkerDispatcher("orochi_workers")
	defer d.Stop()

	proceed := make(chan bool)

	f := func(ctx context.Context) (interface{}, error) {
		<-proceed

		select {
		case <-ctx.Done():
			return nil, errors.Wrap(ctx.Err(), "cancelled")
		default:
			assert.Fail("job is not cancelled correctly")
			return "not cancelled", nil
		}
	}

	tracker := qf(d, f)

	status := tracker.Status()
	assert.False(status.Complete)
	assert.False(status.Success)
	assert.Equal("", status.Error)
	assert.Nil(status.Payload)

	tracker.Stop()
	proceed <- true
	payload, err := tracker.Result()
	assert.Equal(context.Canceled, err)
	assert.Nil(payload)

	jobq.CheckJobStopped(t, tracker)

	status = tracker.Status()
	assert.True(status.Complete)
	assert.False(status.Success)
	assert.Equal(context.Canceled.Error(), status.Error)
	assert.Nil(status.Payload)
	assert.True(status.FinishedAt.After(status.StartedAt))
}

func testQueueStop(t *testing.T, qf queueFunc) {
	tests := []struct {
		worker int
	}{
		{1},
		{2},
		{4},
		{6},
	}
	for _, tt := range tests {
		tt := tt
		t.Run(fmt.Sprintf("When the number of workers is %d", tt.worker), func(t *testing.T) {
			assert := assert.New(t)

			d := jobq.NewWorkerDispatcher("orochi_workers", jobq.ForceWorkerN(tt.worker))

			proceed := make(chan bool)

			tracker := d.QueueFunc(context.Background(), func(ctx context.Context) (interface{}, error) {
				<-proceed

				select {
				case <-ctx.Done():
					return nil, errors.Wrap(ctx.Err(), "cancelled")
				}
			})

			status := tracker.Status()
			assert.False(status.Complete)
			assert.False(status.Success)
			assert.Equal("", status.Error)
			assert.Nil(status.Payload)

			if tt.worker > 0 {
				proceed <- true
			}

			d.Stop()
			jobq.CheckDispatcherStopped(t, d)

			payload, err := tracker.Result()
			assert.Equal(context.Canceled, err)
			assert.Nil(payload)

			jobq.CheckJobStopped(t, tracker)

			status = tracker.Status()
			assert.True(status.Complete)
			assert.False(status.Success)
			assert.Equal(context.Canceled.Error(), status.Error)
			assert.Nil(status.Payload)
		})
	}
}

func testQueueTimeout(t *testing.T, qf queueTimedFunc) {
	assert := assert.New(t)

	d := jobq.NewWorkerDispatcher("orochi_workers")
	defer d.Stop()

	timeout := 100 * time.Millisecond
	slop := timeout * 5 / 10

	f := func(ctx context.Context) (interface{}, error) {
		time.Sleep(2 * timeout)
		select {
		case <-ctx.Done():
			return nil, errors.Wrap(ctx.Err(), "cancelled")
		default:
			assert.Fail("job is not timeout correctly")
			return "not timed out", nil
		}
	}

	tracker := qf(d, f, timeout)

	status := tracker.Status()
	assert.False(status.Complete)
	assert.False(status.Success)
	assert.Equal("", status.Error)
	assert.Nil(status.Payload)

	time.Sleep(2 * timeout)

	payload, err := tracker.Result()
	assert.Equal(context.DeadlineExceeded, err)
	assert.Nil(payload)

	jobq.CheckJobStopped(t, tracker)

	status = tracker.Status()
	assert.True(status.Complete)
	assert.False(status.Success)
	assert.Equal(context.DeadlineExceeded.Error(), status.Error)
	assert.Nil(status.Payload)
	assert.True(status.FinishedAt.After(status.StartedAt))

	// check the execution time
	assert.WithinDuration(status.FinishedAt, status.StartedAt, timeout+slop)
}

func testQueueTimeout_completeBeforeTimeout(t *testing.T, qf queueTimedFunc) {
	assert := assert.New(t)

	d := jobq.NewWorkerDispatcher("orochi_workers")
	defer d.Stop()

	timeout := 10 * time.Second

	f := func(ctx context.Context) (interface{}, error) {
		return "success", nil
	}

	tracker := qf(d, f, timeout)

	payload, err := tracker.Result()
	assert.Equal("success", payload)
	assert.Nil(err)

	jobq.CheckJobStopped(t, tracker)

	status := tracker.Status()
	assert.True(status.Complete)
	assert.True(status.Success)
	assert.Equal("", status.Error)
	assert.Equal("success", status.Payload)
	assert.True(status.FinishedAt.After(status.StartedAt))

	// check the execution time
	assert.WithinDuration(status.FinishedAt, status.StartedAt, timeout)
}
