// Copyright (c) 2023 - 2024. vistart. All rights reserved.
// Use of this source code is governed by Apache-2.0 license
// that can be found in the LICENSE file.

package worker

import (
	"context"
	"errors"
	"log"
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
)

// ExampleTask is an example task.
type ExampleTask struct {
	ID int
}

// Execute implements the Task interface method.
func (t *ExampleTask) Execute(ctx context.Context, args ...interface{}) (interface{}, error) {
	// Simulate some time-consuming task
	result := t.ID * 2
	log.Println(t.ID)
	time.Sleep(1 * time.Second)
	log.Println(t.ID, "finished")
	return result, nil
}

type CanceledTask struct {
	ID int
}

func (t *CanceledTask) Execute(ctx context.Context, args ...interface{}) (interface{}, error) {
	time.Sleep(1 * time.Second)
	return nil, context.Canceled
}

type ErrorTask struct {
	ID int
}

func (t *ErrorTask) Execute(ctx context.Context, args ...interface{}) (interface{}, error) {
	time.Sleep(1 * time.Second)
	return nil, context.DeadlineExceeded
}

type PanicTask struct {
	ID int
}

func (t *PanicTask) Execute(ctx context.Context, args ...interface{}) (interface{}, error) {
	time.Sleep(1 * time.Second)
	panic("task panicked")
}

// TestWorkerPoolBasicFunctionality tests the basic functionality of the worker pool.
func TestWorkerPoolBasicFunctionality(t *testing.T) {
	metrics := &metrics{}
	pool := NewPool(WithMaxWorkers(3), WithWorkerMetrics(metrics))
	defer pool.Shutdown()

	ctx := context.TODO()
	resultChan := pool.Submit(ctx, &ExampleTask{ID: 1})

	result := <-resultChan
	assert.NoError(t, result.Err)
	assert.Equal(t, 2, result.Result)
}

// TestWorkerPoolResize tests resizing the worker pool.
func TestWorkerPoolResize(t *testing.T) {
	t.Run("normal", func(t *testing.T) {
		metrics := &metrics{}
		pool := NewPool(WithMaxWorkers(3), WithWorkerMetrics(metrics))
		defer pool.Shutdown()

		pool.Resize(5, false)
		time.Sleep(time.Millisecond * 100)
		assert.Equal(t, 5, metrics.GetMetrics().CurrentCapacity)

		pool.Resize(2, false)
		time.Sleep(time.Millisecond * 100)
		assert.Equal(t, 2, metrics.GetMetrics().CurrentCapacity)
	})
	t.Run("downsized while all workerWg are busy", func(t *testing.T) {
		metrics := &metrics{}
		pool := NewPool(WithMaxWorkers(3), WithWorkerMetrics(metrics))
		defer pool.Shutdown()

		pool.Submit(context.Background(), &ExampleTask{ID: 1})
		pool.Submit(context.Background(), &ExampleTask{ID: 2})
		pool.Submit(context.Background(), &ExampleTask{ID: 3})

		time.Sleep(time.Millisecond * 100)
		pool.Resize(1, false)
		// If there are busy workerWg, the capacity will not be interrupted when reducing the capacity,
		// and the capacity will not be reduced immediately.
		// The capacity will be reduced only after the workerWg are finished.
		time.Sleep(time.Millisecond * 100)
		m := metrics.GetMetrics()
		assert.Equal(t, 3, m.CurrentCapacity)
		assert.Equal(t, 0, m.IdleWorkers)
		assert.Equal(t, 3, m.WorkingWorkers)
		assert.Equal(t, 0, m.WaitingTasks)
		time.Sleep(time.Second)
		m = metrics.GetMetrics()
		assert.Equal(t, 1, m.CurrentCapacity)
		assert.Equal(t, 1, m.IdleWorkers)
		assert.Equal(t, 0, m.WorkingWorkers)
		assert.Equal(t, 0, m.WaitingTasks)
		assert.Equal(t, 3, m.TotalSubmittedTasks)
		assert.Equal(t, 3, m.TotalCompletedTasks)
		assert.Equal(t, 3, m.SuccessfulTasks)
		assert.Equal(t, 0, m.FailedTasks)
		assert.Equal(t, 0, m.CanceledTasks)
	})
	t.Run("negative", func(t *testing.T) {
		pool := NewPool(WithMaxWorkers(3))
		defer pool.Shutdown()

		err := pool.Resize(-1, false)
		assert.Error(t, err)
		assert.Equal(t, "worker count cannot be negative", err.Error())
	})
}

// TestWorkerPoolRenameWorker tests renaming a worker.
func TestWorkerPoolRenameWorker(t *testing.T) {
	pool := NewPool(WithMaxWorkers(3))
	defer pool.Shutdown()

	err := pool.RenameWorker("worker-1", "renamed-worker-1")
	assert.NoError(t, err)

	err = pool.RenameWorker("worker-1", "renamed-worker-1")
	assert.Error(t, err)
	assert.Equal(t, "new worker ID already exists", err.Error())

	err = pool.RenameWorker("worker-1", "renamed-worker-2")
	assert.Error(t, err)
	assert.Equal(t, "old worker ID not found", err.Error())
}

// TestWorkerPoolExitWorkerByID tests exiting a specific worker by its ID.
func TestWorkerPoolExitWorkerByID(t *testing.T) {
	p := NewPool(WithMaxWorkers(3))
	defer p.Shutdown()

	p.Resize(1, false)
	time.Sleep(time.Millisecond * 100)
	assert.Len(t, p.(*pool).workerMap, 1)
	var name string
	for name = range p.(*pool).workerMap {
		name := name
		err := p.ExitWorkerByID(name, false)
		assert.NoError(t, err)
	}

	err := p.ExitWorkerByID(name, false)
	assert.Error(t, err)
	assert.Equal(t, "worker ID not found", err.Error())

	assert.Equal(t, "", p.(*pool).getWorkerID())
}

type wrongMetrics struct {
}

func (wrongMetrics) GetMetrics() Metrics   { return Metrics{} }
func (wrongMetrics) ResetAllMetrics()      {}
func (wrongMetrics) ResetMetric(int) error { return nil }

// TestWorkerPoolMetrics tests the metrics functionality of the worker pool.
func TestWorkerPoolMetrics(t *testing.T) {
	t.Run("normal", func(t *testing.T) {
		metrics := NewWorkerMetrics()
		pool := NewPool(WithMaxWorkers(3), WithWorkerMetrics(metrics))
		defer pool.Shutdown()

		ctx := context.TODO()
		for i := 0; i < 10; i++ {
			pool.Submit(ctx, &ExampleTask{ID: i})
		}
		time.Sleep(time.Second + time.Millisecond*500)
		assert.Equal(t, 3, metrics.GetMetrics().CurrentCapacity)
		assert.Equal(t, 3, metrics.GetMetrics().WorkingWorkers)
		assert.Equal(t, 0, metrics.GetMetrics().IdleWorkers)

		time.Sleep(3 * time.Second) // Wait for tasks to complete

		assert.Equal(t, 3, metrics.GetMetrics().CurrentCapacity)
		assert.Equal(t, 0, metrics.GetMetrics().WorkingWorkers)
		assert.Equal(t, 3, metrics.GetMetrics().IdleWorkers)
		assert.Equal(t, 0, metrics.GetMetrics().WaitingTasks)
		assert.Equal(t, 10, metrics.GetMetrics().TotalSubmittedTasks)
		assert.Equal(t, 10, metrics.GetMetrics().TotalCompletedTasks)
		assert.Equal(t, 10, metrics.GetMetrics().SuccessfulTasks)
		assert.Equal(t, 0, metrics.GetMetrics().FailedTasks)
		assert.Equal(t, 0, metrics.GetMetrics().CanceledTasks)
	})
	t.Run("wrong type", func(t *testing.T) {
		metrics := &wrongMetrics{}
		pool := NewPool(WithMaxWorkers(3), WithWorkerMetrics(metrics))
		defer pool.Shutdown()
	})
}

// TestWorkerMetricsReset tests the resetting of metrics.
func TestWorkerMetricsReset(t *testing.T) {
	metrics := &metrics{}
	pool := NewPool(WithMaxWorkers(3), WithWorkerMetrics(metrics))
	defer pool.Shutdown()

	ctx := context.TODO()
	for i := 0; i < 5; i++ {
		pool.Submit(ctx, &ExampleTask{ID: i})
	}

	time.Sleep(3 * time.Second) // Wait for tasks to complete

	metrics.ResetMetric(CurrentCapacity)
	assert.Equal(t, 0, metrics.GetMetrics().CurrentCapacity)
	metrics.ResetMetric(WorkingWorkers)
	assert.Equal(t, 0, metrics.GetMetrics().WorkingWorkers)
	metrics.ResetMetric(IdleWorkers)
	assert.Equal(t, 0, metrics.GetMetrics().IdleWorkers)
	metrics.ResetMetric(WaitingTasks)
	assert.Equal(t, 0, metrics.GetMetrics().WaitingTasks)
	metrics.ResetMetric(TotalSubmittedTasks)
	assert.Equal(t, 0, metrics.GetMetrics().TotalSubmittedTasks)
	metrics.ResetMetric(TotalCompletedTasks)
	assert.Equal(t, 0, metrics.GetMetrics().TotalCompletedTasks)
	metrics.ResetMetric(SuccessfulTasks)
	assert.Equal(t, 0, metrics.GetMetrics().SuccessfulTasks)
	metrics.ResetMetric(FailedTasks)
	assert.Equal(t, 0, metrics.GetMetrics().FailedTasks)
	metrics.ResetMetric(CanceledTasks)
	assert.Equal(t, 0, metrics.GetMetrics().CanceledTasks)
	err := metrics.ResetMetric(-1)
	assert.Error(t, err)

	metrics.ResetAllMetrics()
	assert.Equal(t, Metrics{}, metrics.GetMetrics())
}

// TestWorkerPoolCancelTask tests the cancellation of tasks.
func TestWorkerPoolCancelTask(t *testing.T) {
	t.Run("worker returned canceled", func(t *testing.T) {
		metrics := &metrics{}
		pool := NewPool(WithMaxWorkers(1), WithWorkerMetrics(metrics))
		defer pool.Shutdown()

		resultChan := pool.Submit(context.Background(), &CanceledTask{ID: 1})
		result := <-resultChan
		assert.Error(t, result.Err)
		assert.Equal(t, context.Canceled, result.Err)

		time.Sleep(time.Millisecond * 100)
		m := metrics.GetMetrics()
		assert.Equal(t, 1, m.CurrentCapacity)
		assert.Equal(t, 0, m.WorkingWorkers)
		assert.Equal(t, 1, m.IdleWorkers)
		assert.Equal(t, 0, m.WaitingTasks)
		assert.Equal(t, 1, m.TotalSubmittedTasks)
		assert.Equal(t, 1, m.TotalCompletedTasks)
		assert.Equal(t, 0, m.SuccessfulTasks)
		assert.Equal(t, 0, m.FailedTasks)
		assert.Equal(t, 1, m.CanceledTasks)
	})
	t.Run("worker returned error", func(t *testing.T) {
		metrics := &metrics{}
		pool := NewPool(WithMaxWorkers(1), WithWorkerMetrics(metrics))
		defer pool.Shutdown()

		resultChan := pool.Submit(context.Background(), &ErrorTask{ID: 1})
		result := <-resultChan
		assert.Error(t, result.Err)
		assert.Equal(t, context.DeadlineExceeded, result.Err)

		time.Sleep(time.Millisecond * 100)
		m := metrics.GetMetrics()
		assert.Equal(t, 1, m.CurrentCapacity)
		assert.Equal(t, 0, m.WorkingWorkers)
		assert.Equal(t, 1, m.IdleWorkers)
		assert.Equal(t, 0, m.WaitingTasks)
		assert.Equal(t, 1, m.TotalSubmittedTasks)
		assert.Equal(t, 1, m.TotalCompletedTasks)
		assert.Equal(t, 0, m.SuccessfulTasks)
		assert.Equal(t, 1, m.FailedTasks)
		assert.Equal(t, 0, m.CanceledTasks)
	})
	t.Run("worker timeout during working", func(t *testing.T) {
		metrics := &metrics{}
		pool := NewPool(WithMaxWorkers(1), WithWorkerMetrics(metrics))
		defer pool.Shutdown()

		ctx, _ := context.WithTimeout(context.Background(), time.Millisecond*500)
		resultChan := pool.Submit(ctx, &ExampleTask{ID: 1})
		result := <-resultChan
		assert.NoError(t, result.Err)

		time.Sleep(time.Millisecond * 100)
		m := metrics.GetMetrics()
		assert.Equal(t, 1, m.CurrentCapacity)
		assert.Equal(t, 0, m.WorkingWorkers)
		assert.Equal(t, 1, m.IdleWorkers)
		assert.Equal(t, 0, m.WaitingTasks)
		assert.Equal(t, 1, m.TotalSubmittedTasks)
		assert.Equal(t, 1, m.TotalCompletedTasks)
		assert.Equal(t, 0, m.SuccessfulTasks)
		assert.Equal(t, 0, m.FailedTasks)
		assert.Equal(t, 1, m.CanceledTasks)
	})
	t.Run("worker candidate canceled during working", func(t *testing.T) {
		metrics := &metrics{}
		pool := NewPool(WithMaxWorkers(1), WithWorkerMetrics(metrics))
		defer pool.Shutdown()

		ctx, _ := context.WithTimeout(context.Background(), time.Millisecond*500)
		resultChan1 := pool.Submit(ctx, &ExampleTask{ID: 1})
		resultChan2 := pool.Submit(ctx, &ExampleTask{ID: 2})
		resultChan3 := pool.Submit(ctx, &ExampleTask{ID: 3})

		results := []TaskResult{<-resultChan1, <-resultChan2, <-resultChan3}
		// Due to Go's runtime scheduling, the order of task execution is not guaranteed.
		// Therefore, we only ensure that exactly two tasks have a DeadlineExceeded error.
		deadlineExceededCount := 0

		// Count the number of tasks that were canceled due to deadline exceeded
		for _, result := range results {
			if errors.Is(result.Err, context.DeadlineExceeded) {
				deadlineExceededCount++
			}
		}

		// Assert that exactly two tasks were canceled due to deadline exceeded
		assert.Equal(t, 2, deadlineExceededCount, "expected exactly two tasks to be canceled due to deadline exceeded")

		time.Sleep(time.Millisecond * 100)
		m := metrics.GetMetrics()
		assert.Equal(t, 1, m.CurrentCapacity)
		assert.Equal(t, 0, m.WorkingWorkers)
		assert.Equal(t, 1, m.IdleWorkers)
		assert.Equal(t, 0, m.WaitingTasks)
		assert.Equal(t, 3, m.TotalSubmittedTasks)
		assert.Equal(t, 3, m.TotalCompletedTasks)
		assert.Equal(t, 0, m.SuccessfulTasks)
		assert.Equal(t, 0, m.FailedTasks)
		assert.Equal(t, 3, m.CanceledTasks)
	})
}

// TestWorkerPoolHandleConcurrentTasks tests handling of multiple concurrent tasks.
func TestWorkerPoolHandleConcurrentTasks(t *testing.T) {
	metrics := &metrics{}
	pool := NewPool(WithMaxWorkers(5), WithWorkerMetrics(metrics))
	defer pool.Shutdown()

	ctx := context.TODO()
	for i := 0; i < 10; i++ {
		pool.Submit(ctx, &ExampleTask{ID: i})
	}

	time.Sleep(5 * time.Second) // Wait for tasks to complete

	assert.Equal(t, 10, metrics.GetMetrics().TotalCompletedTasks)
	assert.Equal(t, 10, metrics.GetMetrics().SuccessfulTasks)
}

// TestWorkerPoolNoMetrics tests the worker pool without metrics.
func TestWorkerPoolNoMetrics(t *testing.T) {
	pool := NewPool(WithMaxWorkers(3))
	defer pool.Shutdown()

	ctx := context.TODO()
	resultChan := pool.Submit(ctx, &ExampleTask{ID: 1})

	result := <-resultChan
	assert.NoError(t, result.Err)
	assert.Equal(t, 2, result.Result)
}

func TestListWorkers(t *testing.T) {
	p := NewPool(WithMaxWorkers(5))

	// Ensure the initial worker count is as expected
	workers := p.ListWorkers()
	if len(workers) != 5 {
		t.Fatalf("expected 5 workerWg, got %d", len(workers))
	}

	p.Shutdown()
}

func TestStopWorkerByID(t *testing.T) {
	t.Run("all workerWg are idle", func(t *testing.T) {
		p := NewPool(WithMaxWorkers(5))
		defer p.Shutdown()

		// Get the list of workerWg and stop one
		workers := p.ListWorkers()
		if len(workers) == 0 {
			t.Fatal("no workerWg available to stop")
		}
		workerID := workers[0]

		err := p.StopWorkerByID(workerID)
		if err != nil {
			t.Fatalf("failed to stop worker: %v", err)
		}

		// Ensure the worker count is decreased by 1
		remainingWorkers := p.ListWorkers()
		if len(remainingWorkers) != 5 {
			t.Fatalf("expected 5 workerWg, got %d", len(remainingWorkers))
		}

		// Try stopping the non-existent worker, should get an error
		err = p.StopWorkerByID(workerID + "1")
		if err == nil {
			t.Fatal("expected an error when stopping a non-existent worker, got none")
		}
	})
	t.Run("stop 1 worker", func(t *testing.T) {
		p := NewPool(WithMaxWorkers(3))
		defer p.Shutdown()

		p.Submit(context.Background(), &ExampleTask{ID: 1})
		p.Submit(context.Background(), &ExampleTask{ID: 2})
		p.Submit(context.Background(), &ExampleTask{ID: 3})

		time.Sleep(time.Millisecond * 100)
		workers := p.ListWorkers()
		workerID := workers[0]
		err := p.StopWorkerByID(workerID)
		if err != nil {
			t.Fatalf("failed to stop worker: %v", err)
		}
		time.Sleep(time.Millisecond * 100)

		remainingWorkers := p.ListWorkers()
		if len(remainingWorkers) != 3 {
			t.Fatalf("expected 3 workerWg, got %d", len(remainingWorkers))
		}

		err = p.StopWorkerByID(workerID)
		if err != nil {
			t.Fatalf("expected no error when stopping a non-existent worker, got %v", err.Error())
		}
	})
}

func TestExitWorkerByID(t *testing.T) {
	t.Run("all workerWg are idle", func(t *testing.T) {
		p := NewPool(WithMaxWorkers(5))
		defer p.Shutdown()

		// Get the list of workerWg and stop one
		workers := p.ListWorkers()
		if len(workers) == 0 {
			t.Fatal("no workerWg available to stop")
		}
		workerID := workers[0]

		err := p.ExitWorkerByID(workerID, false)
		if err != nil {
			t.Fatalf("failed to exit worker: %v", err)
		}

		err = p.ExitWorkerByID(workerID, false)
		if err == nil {
			t.Fatal("expected an error when exiting a exited worker, got none")
		}

		// Ensure the worker count is decreased by 1
		remainingWorkers := p.ListWorkers()
		if len(remainingWorkers) != 4 {
			t.Fatalf("expected 4 workerWg, got %d", len(remainingWorkers))
		}

		// Try exiting the non-existent worker, should get an error
		err = p.ExitWorkerByID(workerID+"1", false)
		if err == nil {
			t.Fatal("expected an error when exiting a non-existent worker, got none")
		}

		workerID = remainingWorkers[0]
		err = p.ExitWorkerByID(workerID, true)
		if err != nil {
			t.Fatalf("failed to exit worker: %v", err)
		}
	})
	t.Run("exit 1 worker but not stop immediately when all running", func(t *testing.T) {
		p := NewPool(WithMaxWorkers(3))
		defer p.Shutdown()

		p.Submit(context.Background(), &ExampleTask{ID: 1})
		p.Submit(context.Background(), &ExampleTask{ID: 2})
		p.Submit(context.Background(), &ExampleTask{ID: 3})

		time.Sleep(time.Millisecond * 100)
		workers := p.ListWorkers()
		workerID := workers[0]
		err := p.ExitWorkerByID(workerID, false)
		if err != nil {
			t.Fatalf("failed to exit worker: %v", err)
		}
		time.Sleep(time.Millisecond * 100)

		remainingWorkers := p.ListWorkers()
		if len(remainingWorkers) != 2 {
			t.Fatalf("expected 2 workerWg, got %d", len(remainingWorkers))
		}

		err = p.ExitWorkerByID(workerID, true)
		if err == nil {
			t.Fatal("expected an error when stopping a non-existent worker, got none")
		}
	})
	t.Run("exit 1 worker and stop immediately when all running", func(t *testing.T) {
		p := NewPool(WithMaxWorkers(3))
		defer p.Shutdown()

		p.Submit(context.Background(), &ExampleTask{ID: 1})
		p.Submit(context.Background(), &ExampleTask{ID: 2})
		p.Submit(context.Background(), &ExampleTask{ID: 3})

		time.Sleep(time.Millisecond * 100)
		workers := p.ListWorkers()
		workerID := workers[0]
		err := p.ExitWorkerByID(workerID, true)
		if err != nil {
			t.Fatalf("failed to exit worker: %v", err)
		}
		time.Sleep(time.Millisecond * 100)

		remainingWorkers := p.ListWorkers()
		if len(remainingWorkers) != 2 {
			t.Fatalf("expected 2 workerWg, got %d", len(remainingWorkers))
		}

		err = p.ExitWorkerByID(workerID, true)
		if err == nil {
			t.Fatal("expected an error when stopping a non-existent worker, got none")
		}
	})
}

func TestPanicRecovery(t *testing.T) {
	// Create a new worker pool with one worker.
	p := NewPool(WithMaxWorkers(1))

	resultChan := p.Submit(context.Background(), &PanicTask{ID: 1})

	// Wait for the result.
	result := <-resultChan

	// Check if the result contains the expected panic error.
	if result.Err == nil || result.Err.Error() != "task panicked: task panicked" {
		t.Errorf("expected panic error, got %v", result.Err)
	}

	// Shutdown the pool.
	p.Shutdown()
}
