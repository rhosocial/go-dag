# Worker Pool

The `worker` package provides a robust and efficient worker pool implementation. It supports dynamic resizing of the pool, task submission, worker metrics tracking, and graceful shutdown. Additionally, it includes built-in panic recovery to ensure stability during task execution.

## Features

- **Dynamic Resizing**: Adjust the number of workers in the pool dynamically.
- **Task Submission**: Submit tasks for asynchronous execution.
- **Metrics Tracking**: Track various metrics such as successful, failed, and canceled tasks.
- **Graceful Shutdown**: Shutdown the worker pool gracefully, ensuring all tasks are completed.
- **Panic Recovery**: Recover from panics during task execution, ensuring the worker pool remains stable.

## Installation

To install the `worker` package, use the following command:

```shell
go get github.com/rhosocial/go-dag/workflow/standard/worker
```

## Usage

### Creating a Worker Pool

To create a worker pool using the `worker` package, you can use the `NewPool` function along with optional configuration options provided as `Option` functions.

```go
package main

import (
    "context"
    "fmt"
    "time"

    "github.com/rhosocial/go-dag/workflow/standard/worker"
)

type myTask struct {
    name string
}

func (t *myTask) Execute(ctx context.Context, args ...interface{}) (interface{}, error) {
    // Simulate work
    time.Sleep(2 * time.Second)
    return fmt.Sprintf("Task %s completed", t.name), nil
}

func main() {
    // Create a new worker pool with a maximum of 3 workers
    p := worker.NewPool(worker.WithMaxWorkers(3))

    // Use the pool as shown in the examples below

    // Shutdown the worker pool
    p.Shutdown()
}
```

> Note that while `WithMaxWorkers` is optional, it must be provided at least once to ensure the pool has workers.
> If not provided, the pool will be created with zero workers.
> `WithWorkerMetrics` is also optional. If not provided, metrics will not be recorded.

### Defining a Task

Tasks in the `worker` package are defined by implementing the `Task` interface,
which requires the `Execute` method to be implemented.
Here's an example of defining a task:

```go
package main

import "context"

type MyTask struct{}

func (t *MyTask) Execute(ctx context.Context, args ...interface{}) (interface{}, error) {
    // Implement task logic here
}
```

### Submitting a Task

To submit a task to the worker pool for asynchronous execution, use the `Submit` method.
It returns a channel that will receive the result of the task execution.

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

### Recover from Panicked

The `worker` package includes built-in panic recovery to handle unexpected errors during task execution.
If a task panics, the worker will recover from the panic and return an error, ensuring that the worker pool remains stable.

```go
type panicTask struct{}

func (t *panicTask) Execute(ctx context.Context, args ...interface{}) (interface{}, error) {
    panic("something went wrong")
}

resultChan := p.Submit(context.Background(), &panicTask{})

// Retrieve the result
result := <-resultChan
if result.Err != nil {
    fmt.Printf("Task failed: %v\n", result.Err)
}
```

### Resizing the Pool

You can adjust the number of workers in the pool dynamically using the `Resize` method.
This method allows you to increase or decrease the number of workers based on your application's needs.
The second parameter `stopWorker` in Resize method behaves the same way as in the `ExitWorkerByID` method 
and is effective only when reducing the number of workers.

```go
// Increase the number of workers
err := p.Resize(5, false)

// Decrease the number of workers
err = p.Resize(2, true)
```

> Note that if you want to decrease the pool capacity and some of the workers being removed are busy,
> those workers will not stop immediately.
> The resizing will complete only after all such busy workers have finished their tasks.

### Listing Workers

To retrieve the list of worker IDs currently active in the pool, use the `ListWorkers` method.

```go
workerIDs := p.ListWorkers()
fmt.Printf("Active workers: %v\n", workerIDs)
```

### Stopping a Worker by ID

To stop a specific worker by its ID, use the `StopWorkerByID` method.
This method cancels the currently executing task of the worker,
allowing it to receive and execute new tasks afterward.

```go
err := p.StopWorkerByID("worker-1")
if err != nil {
    fmt.Printf("Failed to stop worker: %v\n", err)
}
```

### Exiting a Worker by ID

To terminate a specific worker by its ID, use the `ExitWorkerByID` method.
This method immediately terminates the worker, optionally canceling its current task execution.

```go
err := p.ExitWorkerByID("worker-2", true)
if err != nil {
    fmt.Printf("Failed to exit worker: %v\n", err)
}
```

This method terminates and exits a worker. If `stopWorker` is true, the task's context will be cancelled;
otherwise, it will not be notified. The method immediately removes the worker, even if the task is still running.
Ensure that tasks can handle context's `Done` signal or complete within a reasonable time.

### Using Metrics

To track metrics related to the worker pool, implement the `MetricsProvider` interface and provide it
during the creation of the pool using the `WithWorkerMetrics` option.

```go
type myMetricsProvider struct {
    // Implement metrics fields
}

func (m *myMetricsProvider) GetMetrics() worker.Metrics {
    // Return the current metrics
}

func (m *myMetricsProvider) ResetAllMetrics() {
    // Reset all metrics
}

func (m *myMetricsProvider) ResetMetric(metric int) error {
    // Reset a specific metric
}

metricsProvider := &myMetricsProvider{}

// Create a new worker pool with metrics tracking
p := worker.NewPool(worker.WithMaxWorkers(3), worker.WithWorkerMetrics(metricsProvider))

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