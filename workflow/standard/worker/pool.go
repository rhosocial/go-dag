// Copyright (c) 2023 - 2024. vistart. All rights reserved.
// Use of this source code is governed by Apache-2.0 license
// that can be found in the LICENSE file.

// Package worker provides a robust and efficient worker pool implementation. It supports dynamic resizing of the pool,
// task submission, worker metrics tracking, and graceful shutdown.
// Additionally, it includes built-in panic recovery to ensure stability during task execution.
package worker

import (
	"context"
	"errors"
	"fmt"
	"sync"
)

// Task is the interface that all tasks need to implement.
type Task interface {
	Execute(ctx context.Context, args ...interface{}) (interface{}, error)
}

// Pool is the interface for a worker pool.
type Pool interface {
	// Submit submits a task to the worker pool.
	Submit(ctx context.Context, task Task, args ...interface{}) <-chan TaskResult
	// Resize adjusts the number of workers in the pool.
	Resize(count int, stopWorker bool) error
	// Shutdown shuts down the worker pool and waits for all tasks to complete.
	Shutdown()
	// ListWorkers returns all worker IDs in the pool.
	ListWorkers() []string
	// StopWorkerByID stops a worker by its ID without exiting.
	StopWorkerByID(id string) error
	// ExitWorkerByID exits a specific worker by its ID.
	ExitWorkerByID(id string, stopWorker bool) error
	// RenameWorker renames a specific worker by its ID.
	RenameWorker(oldID, newID string) error
}

// MetricsProvider is the interface for metrics retrieval and management.
type MetricsProvider interface {
	// GetMetrics returns the current metrics.
	GetMetrics() Metrics
	// ResetAllMetrics resets all the metrics to zero.
	ResetAllMetrics()
	// ResetMetric resets a specific metric by name.
	ResetMetric(metric int) error
}

// TaskResult represents the result of a task execution.
type TaskResult struct {
	Result interface{}
	Err    error
}

// workerChannels represents the channels used to control a worker.
type workerChannels struct {
	quit   chan struct{}
	cancel chan error
}

// pool is the concrete implementation of Pool.
type pool struct {
	tasks       chan TaskWithArgs
	workerCount int
	workerWg    sync.WaitGroup
	quit        chan struct{}
	mutex       sync.Mutex
	workerMap   map[string]*workerChannels
	workerIndex int
	metrics     MetricsProvider
	metricsChan chan metricsUpdate
}

// TaskWithArgs represents a task with its arguments and result channel.
type TaskWithArgs struct {
	task       Task
	ctx        context.Context
	cancel     context.CancelCauseFunc
	args       []interface{}
	resultChan chan TaskResult
}

// Option is a function that applies a configuration to the worker pool.
type Option func(*pool)

// WithMaxWorkers sets the maximum number of workers.
//
// This option, although optional, is recommended to specify. It controls the
// upper limit of workers that can be active in the pool at any time. If not specified,
// the pool will not create any workers, effectively being a pool with zero capacity.
//
// Note:
//   - Only the last specified WithMaxWorkers option will take effect if called multiple times.
func WithMaxWorkers(max int) Option {
	return func(p *pool) {
		_ = p.Resize(max, false)
	}
}

// WithWorkerMetrics sets the metrics provider.
//
// This option is optional. If not specified, the worker pool will not track or update
// any metrics related to the pool's operation.
//
// Parameters:
//   - metrics: The metrics provider to use for tracking pool metrics.
func WithWorkerMetrics(metrics MetricsProvider) Option {
	return func(p *pool) {
		p.metrics = metrics
	}
}

// NewPool creates a new worker pool.
func NewPool(options ...Option) Pool {
	pool := &pool{
		tasks:       make(chan TaskWithArgs),
		quit:        make(chan struct{}),
		workerMap:   make(map[string]*workerChannels),
		metricsChan: make(chan metricsUpdate, 100),
	}

	for _, opt := range options {
		opt(pool)
	}

	go pool.handleMetricsUpdates()

	return pool
}

// startWorker creates a new worker with a unique ID and adds it to the worker pool.
//
// This method spawns a goroutine that listens for tasks from the pool's task channel
// and processes them. Each worker is assigned a unique ID and continuously checks for
// incoming tasks. If a quit signal is received and the worker is idle, it exits
// immediately. If the worker is processing a task, it listens for a cancel signal to
// interrupt the task execution.
//
// After completing a task, the method updates pool metrics based on the task outcome:
//   - If the task completes successfully, the SuccessfulTasks metric is incremented.
//   - If the task fails during execution, the FailedTasks metric is incremented.
//   - If the task is canceled due to external reasons, the CanceledTasks metric is incremented.
//
// Additionally, each successful worker startup increments the CurrentCapacity metric,
// while worker exits decrement it. During task execution, the WorkingWorkers metric
// increments, and decrements after task completion.
//
// If a panic occurs during task execution, it is recovered and the FailedTasks metric
// is incremented.
//
// For more detailed information about metrics, refer to the Metrics struct documentation.
//
// Returns:
//   - string: The unique ID assigned to the started worker.
//
// Note:
//   - Workers handle tasks asynchronously, and their execution order is managed by the
//     Go runtime. There is no guarantee that tasks submitted earlier will finish
//     executing before later submissions, or vice versa.
func (p *pool) startWorker() string {
	p.mutex.Lock()
	p.workerIndex++
	workerID := fmt.Sprintf("worker-%d", p.workerIndex)
	channels := &workerChannels{
		quit:   make(chan struct{}),
		cancel: make(chan error),
	}
	p.workerMap[workerID] = channels
	p.mutex.Unlock()

	p.workerWg.Add(1)
	go func(p *pool, id string, channels *workerChannels) {
		defer func() {
			p.workerWg.Done()
			p.metricsChan <- metricsUpdate{deltaCapacity: -1}
		}()
		p.metricsChan <- metricsUpdate{deltaCapacity: 1}
		for {
			select {
			case taskWithArgs := <-p.tasks:
				p.metricsChan <- metricsUpdate{deltaWorking: 1}

				var result TaskResult
				var err error

				// Execute the task asynchronously
				done := make(chan struct{})
				go func() {
					for {
						select {
						case <-done:
							return
						case cause, ok := <-channels.cancel:
							if ok {
								cancel := taskWithArgs.cancel
								if cancel != nil {
									cancel(cause)
								}
							}
							return
						}
					}
				}()
				go func() {
					defer close(done)
					defer func() {
						if r := recover(); r != nil {
							result.Err = fmt.Errorf("task panicked: %v", r)
						}
					}()
					result.Result, err = taskWithArgs.task.Execute(taskWithArgs.ctx, taskWithArgs.args...)
					result.Err = err
				}()

				// Wait for the task to complete or the context to be cancelled
				delta := metricsUpdate{deltaWorking: -1}
				waiting := true
				for waiting {
					select {
					case <-taskWithArgs.ctx.Done():
						// Task was cancelled
						delta.deltaCanceled = 1
						waiting = false
						// Send the empty result back to the submitter
						taskWithArgs.resultChan <- TaskResult{}
					case <-done:
						// Task completed
						if result.Err != nil {
							if errors.Is(result.Err, context.Canceled) {
								delta.deltaCanceled = 1
							} else {
								delta.deltaFailed = 1
							}
						} else {
							delta.deltaSuccessful = 1
						}
						waiting = false
						// Send the task result back to the submitter
						taskWithArgs.resultChan <- result
					}
				}
				p.metricsChan <- delta
				close(taskWithArgs.resultChan) // Close the resultChan after sending the result
			case <-channels.cancel:
				continue
			case <-channels.quit:
				// Quit signal received
				// If worker is idle, exit immediately
				return
			}
		}
	}(p, workerID, channels)

	return workerID
}

// Resize adjusts the number of workers in the pool.
//
// This method changes the number of workers in the pool to the specified count.
// If the count is greater than the current number of workers, additional workers
// will be started. If the count is less than the current number of workers, a
// random selection of workers will be chosen to exit.
//
// The stopWorker parameter indicates whether to notify a busy worker to cancel
// its current task before exiting.
//
// Parameters:
//   - count: The desired number of workers after resizing.
//   - stopWorker: If true, notifies busy workers to cancel their tasks before exiting.
//
// Returns:
//   - error: An error if the count is negative.
//
// Usage:
//
//	err := pool.Resize(10, true)
func (p *pool) Resize(count int, stopWorker bool) error {
	if count < 0 {
		return errors.New("worker count cannot be negative")
	}

	p.mutex.Lock()
	currentCount := p.workerCount
	p.workerCount = count
	p.mutex.Unlock()

	if count > currentCount {
		for i := currentCount; i < count; i++ {
			p.startWorker()
		}
	} else if count < currentCount {
		for i := currentCount; i > count; i-- {
			workerID := p.getWorkerID()
			if workerID != "" {
				_ = p.ExitWorkerByID(workerID, stopWorker)
			}
		}
	}
	return nil
}

// getWorkerID returns any worker ID from the workerMap.
func (p *pool) getWorkerID() string {
	p.mutex.Lock()
	defer p.mutex.Unlock()
	for id := range p.workerMap {
		return id
	}
	return ""
}

// Submit submits a task to the worker pool.
//
// The ctx parameter is used to create a derived context with context.WithCancelCause,
// which is then passed to the task's Execute method. The derived context's cancel function
// can be used to cancel the task execution and provide a reason for cancellation.
// The args parameter is passed as is to the task.
//
// This method executes tasks asynchronously, returning a channel that will receive
// the task result upon completion. The channel's result usage is described in the TaskResult comments.
//
// Note: The execution order of tasks by multiple workers is determined by Go's runtime.
// This means there is no guarantee that tasks will be executed in the order they are submitted,
// nor is there a guarantee that tasks with identical logic will complete in the same order
// they were started. Therefore, you should not rely on the submission order to determine
// the order in which tasks finish.
//
// Parameters:
//   - ctx: The context used to derive a new context passed to the task's Execute method.
//   - task: The task to be executed.
//   - args: Additional arguments to be passed to the task.
//
// Returns a channel that will receive the TaskResult upon completion.
func (p *pool) Submit(ctx context.Context, task Task, args ...interface{}) <-chan TaskResult {
	resultChan := make(chan TaskResult, 1)
	p.metricsChan <- metricsUpdate{deltaSubmitted: 1, deltaWaiting: 1}
	ctx_, cancel := context.WithCancelCause(ctx)

	go func() {
		for {
			select {
			case p.tasks <- TaskWithArgs{task: task, ctx: ctx_, cancel: cancel, args: args, resultChan: resultChan}:
				// Task successfully sent to workers
				p.metricsChan <- metricsUpdate{deltaWaiting: -1}
				return
			case <-ctx.Done():
				// When the pool has no idle workers, i.e., p.tasks temporarily stops accepting channel inputs,
				// this branch is executed if the task is to be canceled before being able to be sent to a worker.
				p.metricsChan <- metricsUpdate{deltaWaiting: -1, deltaCanceled: 1}
				resultChan <- TaskResult{Err: ctx.Err()}
				close(resultChan)
				return
			}
		}
	}()
	return resultChan
}

// ExitWorkerByID exits a specific worker by its ID.
//
// This method terminates a worker that is currently executing and exits it.
// If stopWorker is true, the worker's context will be canceled, and the task execution
// will be notified of the cancellation. If stopWorker is false, no cancellation
// notification will be sent.
//
// Parameters:
//   - id: The ID of the worker to be exited.
//   - stopWorker: A boolean flag indicating whether to cancel the task's context.
//
// If the specified ID exists, the worker will be exited immediately. Subsequent calls to this method
// with the same ID will return an error indicating that the worker ID does not exist, even if the
// corresponding task is still running. Therefore, tasks should be capable of responding to the context's
// Done signal or ensuring that they complete execution within a reasonable time frame.
//
// Returns an error if the worker ID is not found.
func (p *pool) ExitWorkerByID(id string, stopWorker bool) error {
	p.mutex.Lock()
	defer p.mutex.Unlock()

	if channels, exists := p.workerMap[id]; exists {
		if stopWorker {
			channels.cancel <- errors.New("worker exited")
		}
		close(channels.quit)
		delete(p.workerMap, id)
		return nil
	}
	return errors.New("worker ID not found")
}

// StopWorkerByID stops a worker by its ID.
//
// This method cancels the currently executing worker, allowing it to receive
// and execute new tasks afterward. If the specified worker is idle,
// nothing will happen.
//
// Parameters:
// - id: The unique identifier of the worker to stop.
//
// Returns an error if the worker ID is not found.
func (p *pool) StopWorkerByID(id string) error {
	p.mutex.Lock()
	defer p.mutex.Unlock()

	if channels, exists := p.workerMap[id]; exists {
		channels.cancel <- errors.New("user canceled")
		return nil
	}
	return errors.New("worker ID not found")
}

// RenameWorker renames a specific worker by its ID.
//
// This method changes the ID of a worker to a new ID.
//
//   - oldID: The current ID of the worker.
//   - newID: The new ID to assign to the worker.
//
// Returns an error if the old ID is not found or the new ID already exists.
func (p *pool) RenameWorker(oldID, newID string) error {
	p.mutex.Lock()
	defer p.mutex.Unlock()

	if _, exists := p.workerMap[newID]; exists {
		return errors.New("new worker ID already exists")
	}

	if channels, exists := p.workerMap[oldID]; exists {
		p.workerMap[newID] = channels
		delete(p.workerMap, oldID)
		return nil
	}
	return errors.New("old worker ID not found")
}

// Shutdown cancels all currently executing workers and immediately closes the pool.
//
// This method cancels all workers that are currently executing tasks and
// immediately shuts down the worker pool without waiting for the completion
// of those tasks.
//
// The pool will be completely shut down, and no new tasks will be accepted.
func (p *pool) Shutdown() {
	p.mutex.Lock()
	for id, channels := range p.workerMap {
		channels.cancel <- errors.New("pool shutdown")
		close(channels.quit)
		delete(p.workerMap, id)
	}
	close(p.quit)
	p.mutex.Unlock()
	p.workerWg.Wait()
}

// ListWorkers returns all worker IDs in the pool.
//
// This method returns a list of IDs for all workers currently in the pool.
//
// Returns a slice of worker IDs.
func (p *pool) ListWorkers() []string {
	p.mutex.Lock()
	defer p.mutex.Unlock()

	ids := make([]string, 0, len(p.workerMap))
	for id := range p.workerMap {
		ids = append(ids, id)
	}
	return ids
}

// handleMetricsUpdates processes metrics updates from the metricsChan.
//
// This method runs in a separate goroutine to handle updates to the metrics of the worker pool.
// If no metrics provider was specified during instantiation, metrics will not be updated.
//
// It listens for metrics updates on the metricsChan and applies the updates to the current
// metrics state. The metrics updates include changes to the number of working workers,
// idle workers, waiting tasks, and other task-related metrics.
//
// This method uses a mutex to ensure that metrics updates are thread-safe.
func (p *pool) handleMetricsUpdates() {
	for update := range p.metricsChan {
		if p.metrics == nil {
			continue
		}

		metrics, ok := p.metrics.(*metrics)
		if !ok {
			continue
		}
		metrics.Lock()
		metrics.data.CurrentCapacity += update.deltaCapacity
		metrics.data.WorkingWorkers += update.deltaWorking
		metrics.data.IdleWorkers = metrics.data.CurrentCapacity - metrics.data.WorkingWorkers
		metrics.data.WaitingTasks += update.deltaWaiting
		metrics.data.TotalSubmittedTasks += update.deltaSubmitted
		metrics.data.SuccessfulTasks += update.deltaSuccessful
		metrics.data.FailedTasks += update.deltaFailed
		metrics.data.CanceledTasks += update.deltaCanceled
		metrics.data.TotalCompletedTasks += update.deltaSuccessful + update.deltaFailed + update.deltaCanceled
		metrics.Unlock()
	}
}
