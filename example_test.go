package jobq_test

import (
	"context"
	"fmt"
	"log"
	"math/rand"
	"time"

	"github.com/honestbee/jobq"
)

func Example_basic() {
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

func Example_timeout() {
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

func Example_getAllResults() {
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
