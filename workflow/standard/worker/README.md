# worker

`worker` is a package providing a robust and efficient worker pool implementation.
It supports dynamic resizing of the pool, task submission, worker metrics tracking, and graceful shutdown.

## Features

- Dynamic resizing of the worker pool
- Task submission with results
- Worker-specific operations (stop, exit, rename)
- Metrics tracking and retrieval
- Graceful shutdown of the pool

## Installation

To install the package, run:

```shell
go get github.com/rhosocial/go-dag/workflow/standard/worker
```

## Usage

### Creating a Worker Pool

To create a new worker pool, use the `NewPool` function with optional configuration options:

```go
package main

import (
    "github.com/rhosocial/go-dag/workflow/standard/worker"
)

func main() {
    pool := worker.NewPool(
        worker.WithMaxWorkers(5),
        worker.WithWorkerMetrics(worker.NewWorkerMetrics()),
    )
    defer pool.Shutdown()
}

```

> Note that while `WithMaxWorkers` is optional, it must be provided at least once to ensure the pool has workers.
> If not provided, the pool will be created with zero workers.
> `WithWorkerMetrics` is also optional. If not provided, metrics will not be recorded.

### Defining a Task

Tasks need to implement the `Task` interface:

```go
package main

import "context"

type MyTask struct{}

func (t *MyTask) Execute(ctx context.Context, args ...interface{}) (interface{}, error) {
    // Task execution logic
    return nil, nil
}
```

### Submitting a Task

Submit a task to the pool using the `Submit` method:

```go
task := &MyTask{}
resultChan := pool.Submit(context.Background(), task, "arg1", "arg2")

result := <-resultChan
if result.Err != nil {
    fmt.Println("Task failed:", result.Err)
} else {
    fmt.Println("Task succeeded:", result.Result)
}
```

### Resizing the Pool

Adjust the number of workers in the pool:

```go
err := pool.Resize(10)
if err != nil {
    fmt.Println("Resize error:", err)
}
```

> Note that if you want to decrease the pool capacity and some of the workers being removed are busy,
> those workers will not stop immediately.
> The resizing will complete only after all such busy workers have finished their tasks.

### Listing Workers

Retrieve the list of worker IDs:

```go
workerIDs := pool.ListWorkers()
fmt.Println("Worker IDs:", workerIDs)
```

### Stopping a Worker by ID

Stop a worker by its ID without removing it from the pool:

```go
err := pool.StopWorkerByID("worker-1")
if err != nil {
    fmt.Println("StopWorkerByID error:", err)
}
```

### Exiting a Worker by ID

```go
err := pool.ExitWorkerByID("worker-1", true)
if err != nil {
    fmt.Println("Exit error:", err)
}
```

This method terminates and exits a worker. If `stopWorker` is true, the task's context will be cancelled;
otherwise, it will not be notified. The method immediately removes the worker, even if the task is still running.
Ensure that tasks can handle context's `Done` signal or complete within a reasonable time.

### Metrics

If you specified `WithWorkerMetrics`, you can retrieve metrics:

```go
metrics := metricsProvider.GetMetrics()
fmt.Printf("Current Capacity: %d\n", metrics.CurrentCapacity)
fmt.Printf("Working Workers: %d\n", metrics.WorkingWorkers)
fmt.Printf("Idle Workers: %d\n", metrics.IdleWorkers)
fmt.Printf("Waiting Tasks: %d\n", metrics.WaitingTasks)
fmt.Printf("Total Submitted Tasks: %d\n", metrics.TotalSubmittedTasks)
fmt.Printf("Total Completed Tasks: %d\n", metrics.TotalCompletedTasks)
fmt.Printf("Successful Tasks: %d\n", metrics.SuccessfulTasks)
fmt.Printf("Failed Tasks: %d\n", metrics.FailedTasks)
fmt.Printf("Canceled Tasks: %d\n", metrics.CanceledTasks)
```

> Note: There may be a delay in metrics updates. Immediately calling `GetMetrics` after performing operations
> on the pool may not reflect the latest state.

To reset metrics:

```go
metricsProvider.ResetAllMetrics()

// Or reset specific metrics
err := metricsProvider.ResetMetric(worker.TotalSubmittedTasks)
if err != nil {
    fmt.Println("Reset metric error:", err)
}
```

> Note:
> The `Metrics` struct should be used for a single worker pool only and is not recommended for reuse across multiple pools.