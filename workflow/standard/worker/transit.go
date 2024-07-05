package worker

import (
	"context"
	"errors"
	"sync/atomic"

	"github.com/rhosocial/go-dag/workflow/standard/channel"
)

// Transit is an implementation of the Transit interface that contains a pool of workers.
type Transit struct {
	ListeningChannels map[string][]string
	SendingChannels   map[string][]string
	worker            channel.WorkerFunc
	workerSubmitter   channel.WorkerFunc
	pool              Pool
	channel.Node
}

// GetWorker returns the worker submitter of the pool.
func (t *Transit) GetWorker() channel.WorkerFunc {
	return t.workerSubmitter
}

// AppendListeningChannels appends a listening channel name(s) to the current node.
//
// key refers to the name of the listening node, and value refers to the channel name of the node.
func (t *Transit) AppendListeningChannels(key string, value ...string) {
	t.ListeningChannels[key] = append(t.ListeningChannels[key], value...)
}

// AppendSendingChannels appends a sending channel name(s) to the current node.
//
// key refers to the name of the sending node, and value refers to the channel name of the node.
func (t *Transit) AppendSendingChannels(key string, value ...string) {
	t.SendingChannels[key] = append(t.SendingChannels[key], value...)
}

var SimpleTaskID int64 = 0

// SimpleTask represents a task with a unique integer ID.
type SimpleTask struct {
	ID     int64
	worker channel.WorkerFunc
}

func (t *SimpleTask) Execute(ctx context.Context, a ...any) (any, error) {
	t.ID = atomic.AddInt64(&SimpleTaskID, 1)
	return t.worker(ctx, a...)
}

// NewTransit creates a new Transit with the given name, incoming channels, and outgoing channels.
//
// Note that worker cannot be nil, and the pool cannot be nil, otherwise an error will be reported.
func NewTransit(name string, incoming, outgoing []string, worker channel.WorkerFunc, pool Pool) (channel.Transit, error) {
	if worker == nil {
		return nil, errors.New("worker is nil")
	}
	if pool == nil {
		return nil, errors.New("pool is nil")
	}
	t := &Transit{
		Node:              channel.NewNode(name, incoming, outgoing),
		ListeningChannels: make(map[string][]string),
		SendingChannels:   make(map[string][]string),
		worker:            worker,
		pool:              pool,
	}
	t.workerSubmitter = func(ctx context.Context, a ...any) (any, error) {
		result := t.pool.Submit(ctx, &SimpleTask{
			worker: worker,
		}, a...)
		output := <-result
		return output.Result, output.Err
	}
	return t, nil
}
