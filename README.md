# jobq

---
**jobq** is a generic worker pool library with dynamic adjust worker number

## Install

```bash
go get github.com/honestbee/jobq
```

## Run The Test

```bash
# under the jobq repo
export GO111MODULE=on
go mod vendor
go test -v -race ./...
```

## Examples

### basic usage

```golang
package main

import (
        "time"

        "github.com/honestbee/jobq"
)

func main() {
        dispatcher := jobq.NewWorkerDispatcher()
        defer dispatcher.Stop()

        job := func(ctx context.Context) (interface{}, error) {
                time.Sleep(200 *time.Millisecond)
                return "success", nil
        }

        tracker := dispatcher.QueueFunc(job)

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
```

### enable dynamic worker

```golang
package main

import (
        "time"

        "github.com/honestbee/jobq"
)

func main() {
        dispatcher := jobq.NewWorkerDispatcher(
                // enable dynamic workers
                jobq.EnableDynamicAdjustWorkers(true),
                // every 1 second will checking the adjustment needs or not
                jobq.WorkerAdjustPeriod(time.Second),
                // set up one of the adjustment condition,
                // when the worker loading reach 80 percentage 
                // will raise the worker number
                jobq.WorkerLoadingBoundPercentage(20, 80),
                // set 10 times worker number when increase
                jobq.WorkerMargin(10),
                // init 1 workers
                jobq.WorkerN(1),
                // init worker pool size to 10
                jobq.WorkerPoolSize(10),
        )
        defer dispatcher.Stop()

        job := func(ctx context.Context) (interface{}, error) {
                time.Sleep(20 *time.Second)
                return "success", nil
        }

        for i := 0; i < 9; i++ {
                // fill up sleeping jobs to let worker loading
                // over 80 percentage
                dispatcher.QueueFunc(job)
        }

        // now the worker number will increase from 1 to 10
}
```

### metrics report

```golang
package main

import (
        "time"

        "github.com/honestbee/jobq"
        "github.com/prometheus/client_golang/prometheus"
)

func main() {
        // some prometeus register
        busyWorker := prometheus.NewGaugeVec(prometheus.GaugeOpts{
                Name: "workerpool_total_workers",
                Help: "Total number of workers in the worker pool.",
        }, []string{"name"}))
        prometheus.MustRegister(busyWorker)

        dispatcher := jobq.NewWorkerDispatcher(
                // set the report function
                jobq.EnableTrackReport(func(p jobq.TrackParams){
                       busyWorker.Set(p.BusyWorkers) 
                }),
                // set every 30 second will report the metrics
                jobq.MetricsReportPeriod(30*time.Second),
        )
        defer dispatcher.Stop()

        // do some job
        // ...
}
```

more [examples](./example_test.go)

## Contribution

please feel free to do the pull request !!!
