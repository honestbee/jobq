package jobq_test

import (
	"context"
	"fmt"
	"log"
	"math/rand"
	"sync"
	"time"

	"github.com/honestbee/jobq"
)

func ExampleQueueFunc() {
	dispatcher := jobq.NewWorkerDispatcher()
	defer dispatcher.Stop()

	tracker := dispatcher.QueueFunc(context.Background(), func(ctx context.Context) (interface{}, error) {
		time.Sleep(200 * time.Millisecond)
		return "success", nil
	})

	payload, err := tracker.Result()
	status := tracker.Status()
	fmt.Printf("complete=%t\n", status.Complete)
	fmt.Printf("success=%t\n", status.Success)
	if err != nil {
		fmt.Printf("err=%s\n", err.Error())
	} else {
		fmt.Printf("payload=%s\n", payload)
	}

	// Output:
	// complete=true
	// success=true
	// payload=success
}

type runner struct {
	sleepDuration time.Duration
}

func (r *runner) Run(ctx context.Context) (interface{}, error) {
	time.Sleep(r.sleepDuration)
	return "success", nil
}

func ExampleQueue() {
	dispatcher := jobq.NewWorkerDispatcher()
	defer dispatcher.Stop()

	job := &runner{
		sleepDuration: 200 * time.Millisecond,
	}
	tracker := dispatcher.Queue(context.Background(), job)

	payload, err := tracker.Result()
	status := tracker.Status()
	fmt.Printf("complete=%t\n", status.Complete)
	fmt.Printf("success=%t\n", status.Success)
	if err != nil {
		fmt.Printf("err=%s\n", err.Error())
	} else {
		fmt.Printf("payload=%s\n", payload)
	}

	// Output:
	// complete=true
	// success=true
	// payload=success
}

func ExampleQueueTimedFunc() {
	dispatcher := jobq.NewWorkerDispatcher()
	defer dispatcher.Stop()

	tracker := dispatcher.QueueTimedFunc(context.Background(), func(ctx context.Context) (interface{}, error) {
		time.Sleep(1 * time.Second)
		return "success", nil
	}, 200*time.Millisecond)

	payload, err := tracker.Result()
	status := tracker.Status()
	fmt.Printf("complete=%t\n", status.Complete)
	fmt.Printf("success=%t\n", status.Success)
	if err != nil {
		fmt.Printf("err=%s\n", err.Error())
	} else {
		fmt.Printf("payload=%s\n", payload)
	}

	// Output:
	// complete=true
	// success=false
	// err=context deadline exceeded
}

func ExampleQueueTimed() {
	dispatcher := jobq.NewWorkerDispatcher()
	defer dispatcher.Stop()

	job := &runner{
		sleepDuration: time.Second,
	}

	tracker := dispatcher.QueueTimed(context.Background(), job, 200*time.Millisecond)
	payload, err := tracker.Result()
	status := tracker.Status()
	fmt.Printf("complete=%t\n", status.Complete)
	fmt.Printf("success=%t\n", status.Success)
	if err != nil {
		fmt.Printf("err=%s\n", err.Error())
	} else {
		fmt.Printf("payload=%s\n", payload)
	}

	// Output:
	// complete=true
	// success=false
	// err=context deadline exceeded
}

func ExampleGetAllResults() {
	dispatcher := jobq.NewWorkerDispatcher()
	defer dispatcher.Stop()

	rand.Seed(time.Now().UnixNano())

	// Queue a normal job func which doesn't get time-out.
	var trackers []jobq.JobTracker
	for i := 0; i < 10; i++ {
		id := i
		tracker := dispatcher.QueueFunc(context.Background(), func(ctx context.Context) (interface{}, error) {
			time.Sleep(time.Duration(rand.Intn(9)+1) * 100 * time.Millisecond)
			return fmt.Sprintf("Job #%d", id), nil
		})
		trackers = append(trackers, tracker)
	}

	for _, tracker := range trackers {
		payload, err := tracker.Result()
		if err != nil {
			log.Fatal(err)
		}
		fmt.Println(payload)
	}

	// Output:
	// Job #0
	// Job #1
	// Job #2
	// Job #3
	// Job #4
	// Job #5
	// Job #6
	// Job #7
	// Job #8
	// Job #9
}

func ExampleReportMetrics() {
	done := make(chan struct{})

	dispatcher := jobq.NewWorkerDispatcher(
		// set the report function
		jobq.EnableTrackReport(func(p jobq.TrackParams) {
			fmt.Println("JobQueueSize:", p.JobQueueSize) // current queued jobs
			fmt.Println("TotalWorkers:", p.TotalWorkers)
			fmt.Println("BusyWorkers:", p.BusyWorkers)
			close(done)
			// using the prometheus here or something else monitoring tools
		}),
		// set every second will report the metrics
		jobq.MetricsReportPeriod(time.Second),
	)
	defer dispatcher.Stop()

	// do some job
	// ...
	<-done

	// Output:
	// JobQueueSize: 0
	// TotalWorkers: 1024
	// BusyWorkers: 0
}

func ExampleDynamicAdjustWorkerNum() {
	dispatcher := jobq.NewWorkerDispatcher(
		// init 1 workers
		jobq.WorkerN(1),
		// init worker pool size to 10
		jobq.WorkerPoolSize(10),
		// enable dynamic workers
		jobq.EnableDynamicAdjustWorkers(true),
		// set every second will checking the adjustment needs or not
		jobq.WorkerAdjustPeriod(time.Second),
		// set up one of the adjustment condition,
		// when the worker loading reach 80 percentage
		// will raise the worker number
		jobq.WorkerLoadingBoundPercentage(20, 80),
		// set 2 times worker number when increase
		jobq.WorkerMargin(2),
	)
	defer dispatcher.Stop()

	done := make(chan struct{})
	wg := new(sync.WaitGroup)
	job := func(ctx context.Context) (interface{}, error) {
		defer wg.Done()
		<-done
		return "success", nil
	}

	for i := 0; i < 9; i++ {
		wg.Add(1)
		// fill up blocking jobs to let worker loading
		// over 80 percentage
		dispatcher.QueueFunc(context.Background(), job)
	}

	// now the worker num will increasing.
	<-done
	wg.Wait()
}
