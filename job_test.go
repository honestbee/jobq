package jobq_test

import (
	"context"
	"testing"
	"time"

	"github.com/stretchr/testify/assert"

	"github.com/honestbee/orochi/internal/jobq"
)

func TestJobTimeoutBeforeIntoWorker(t *testing.T) {
	assert := assert.New(t)

	d := jobq.NewWorkerDispatcher("orochi_workers", jobq.WorkerN(1), jobq.WorkerPoolSize(1))
	defer d.Stop()

	timeout := 100 * time.Millisecond
	slop := timeout * 5 / 10

	// stock a job for the only worker
	d.QueueFunc(context.Background(), func(ctx context.Context) (interface{}, error) {
		time.Sleep(2 * timeout)
		return "success", nil
	})

	inWorkerC := make(chan struct{})
	tracker := d.QueueTimedFunc(context.Background(), func(ctx context.Context) (interface{}, error) {
		defer func() { inWorkerC <- struct{}{} }()
		return "success", nil
	}, timeout)

	select {
	case <-inWorkerC:
		assert.Fail("job goes into worker not timeout first")
	case <-tracker.Done():
		status := tracker.Status()
		assert.True(status.Complete)
		assert.False(status.Success)
		assert.Equal(context.DeadlineExceeded.Error(), status.Error)
		assert.Nil(status.Payload)
		assert.True(status.FinishedAt.After(status.StartedAt))
		assert.WithinDuration(status.FinishedAt, status.StartedAt, timeout+slop)
	}
}
