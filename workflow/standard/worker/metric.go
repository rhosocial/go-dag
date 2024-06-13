// Copyright (c) 2023 - 2024. vistart. All rights reserved.
// Use of this source code is governed by Apache-2.0 license
// that can be found in the LICENSE file.

package worker

import (
	"errors"
	"sync"
)

// Metrics represents the various metrics for the worker pool.
//
// Note that the instance should be used for a single worker pool only and is not recommended for reuse across multiple pools.
type Metrics struct {
	CurrentCapacity     int // Current number of workerWg in the pool
	WorkingWorkers      int // Number of workerWg currently working
	IdleWorkers         int // Number of idle workerWg
	WaitingTasks        int // Number of tasks waiting to be executed
	TotalSubmittedTasks int // Total number of submitted tasks
	TotalCompletedTasks int // Total number of completed tasks
	SuccessfulTasks     int // Number of successfully completed tasks
	FailedTasks         int // Number of failed tasks
	CanceledTasks       int // Number of canceled tasks
}

// Constants for ResetMetric method
const (
	CurrentCapacity = iota
	WorkingWorkers
	IdleWorkers
	WaitingTasks
	TotalSubmittedTasks
	TotalCompletedTasks
	SuccessfulTasks
	FailedTasks
	CanceledTasks
)

// metricsUpdate represents a request to update metrics.
type metricsUpdate struct {
	deltaCapacity   int
	deltaSubmitted  int
	deltaSuccessful int
	deltaFailed     int
	deltaCanceled   int
	deltaWorking    int
	deltaWaiting    int
}

// metrics implements MetricsProvider.
type metrics struct {
	sync.Mutex
	data Metrics
}

// GetMetrics returns the current metrics.
func (m *metrics) GetMetrics() Metrics {
	m.Lock()
	defer m.Unlock()

	idleWorkers := m.data.CurrentCapacity - m.data.WorkingWorkers
	m.data.IdleWorkers = idleWorkers

	return m.data
}

// ResetAllMetrics resets all the metrics to zero.
func (m *metrics) ResetAllMetrics() {
	m.Lock()
	defer m.Unlock()
	m.data = Metrics{}
}

// ResetMetric resets a specific metric by name.
func (m *metrics) ResetMetric(metric int) error {
	m.Lock()
	defer m.Unlock()
	switch metric {
	case CurrentCapacity:
		m.data.CurrentCapacity = 0
	case WorkingWorkers:
		m.data.WorkingWorkers = 0
	case IdleWorkers:
		m.data.IdleWorkers = 0
	case WaitingTasks:
		m.data.WaitingTasks = 0
	case TotalSubmittedTasks:
		m.data.TotalSubmittedTasks = 0
	case TotalCompletedTasks:
		m.data.TotalCompletedTasks = 0
	case SuccessfulTasks:
		m.data.SuccessfulTasks = 0
	case FailedTasks:
		m.data.FailedTasks = 0
	case CanceledTasks:
		m.data.CanceledTasks = 0
	default:
		return errors.New("unknown metric name")
	}
	return nil
}

// NewWorkerMetrics instantiate a worker metrics.
func NewWorkerMetrics() MetricsProvider {
	return &metrics{}
}
