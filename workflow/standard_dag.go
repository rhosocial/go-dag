package workflow

import (
	"context"
	"log"
	"sync"
)

type StandardDAGInterface[TInput, TOutput any] interface {
	SimpleDAGInitInterface
	BuildWorkflow(ctx context.Context) error
	BuildWorkflowInput(result any, inputs ...string)
	BuildWorkflowOutput(ctx context.Context, outputs ...string) *[]any
	CloseWorkflow()
	Execute(ctx context.Context, input *TInput) *TOutput
	RunOnce(ctx context.Context, input *TInput) *TOutput
}

type StandardDAGChannel struct {
	Prefix string
	// channels stores all channels of this directed acyclic graph. The key of the map is the channel channels.
	channels      map[string]chan any
	channelInput  string
	channelOutput string
}

// StandardDAG supports：
// 1. Pipeline execution, track each execution(including progress, elapsed time and logs), record critical path and topology sorting, detect loops, support deadline or time limit.
// 2. StandardDAGWorkflowTransit supports WorkerPool, which can increase or decrease the number of workers as needed, as well as the priority of each worker.
// 3. Node worker supports stateful and stateless. The execution result of the node worker supports the caching mechanism, that is, the same input directly gives the corresponding output instead of executing it once. This feature is limited to stateless worker nodes.
type StandardDAG[TInput, TOutput any] struct {
	workflowTransits []*SimpleDAGWorkflowTransit
	logger           *log.Logger
	StandardDAGChannel
	StandardDAGInterface[TInput, TOutput]
}

func NewStandardDAG[TInput, TOutput any]() *StandardDAG[TInput, TOutput] {
	return &StandardDAG[TInput, TOutput]{
		logger: log.Default(),
	}
}

func (d *StandardDAG[TInput, TOutput]) BuildWorkflow(ctx context.Context) error {
	// Build Input
	if len(d.channelInput) == 0 {
		return ErrChannelInputEmpty
	}

	// Build Output
	if len(d.channelOutput) == 0 {
		return ErrChannelOutputEmpty
	}

	// Build Transits
	if len(d.workflowTransits) == 0 {
		return nil
	}
	for _, t := range d.workflowTransits {
		for _, name := range t.channelInputs {
			if _, existed := d.channels[name]; !existed {
				return ErrChannelNotExist
			}
		}
		for _, name := range t.channelOutputs {
			if _, existed := d.channels[name]; !existed {
				return ErrChannelNotExist
			}
		}
	}
	for _, t := range d.workflowTransits {
		go func(t *SimpleDAGWorkflowTransit) {
			results := d.BuildWorkflowOutput(ctx, t.channelInputs...)
			var result, err = t.worker(ctx, *results...)
			if err != nil {
				d.logger.Printf("worker[%s] error(s) occurred: %s\n", t.name, err.Error())
			}
			d.BuildWorkflowInput(result, t.channelOutputs...)
		}(t)
	}
	return nil
}

func (d *StandardDAG[TInput, TOutput]) CloseWorkflow() {
	if len(d.channelInput) == 0 {
		return
	}
	close(d.channels[d.channelInput])
	if len(d.channelOutput) == 0 {
		return
	}
	//close(d.channelInputs[d.channelOutput])
	if len(d.workflowTransits) == 0 {
		return
	}
	for _, t := range d.workflowTransits {
		for _, c := range t.channelOutputs {
			close(d.channels[c])
		}
	}
}

func (d *StandardDAG[TInput, TOutput]) Execute(ctx context.Context, input *TInput) *TOutput {
	err := d.BuildWorkflow(ctx)
	if err != nil {
		return nil
	}
	var results *TOutput
	var wg sync.WaitGroup
	wg.Add(1)
	go func() {
		defer wg.Done()
		r := d.BuildWorkflowOutput(ctx, d.channelOutput)
		if r, ok := (*r)[0].(TOutput); !ok {
			t := new(TOutput)
			var e = SimpleDAGValueTypeError{actual: r, expect: t}
			d.logger.Println(e)
			results = nil
			return
		} else {
			results = &r
		}
	}()
	d.BuildWorkflowInput(*input, d.channelInput)
	wg.Wait()
	return results
}

func (d *StandardDAG[TInput, TOutput]) RunOnce(ctx context.Context, input *TInput) *TOutput {
	defer d.CloseWorkflow()
	return d.Execute(ctx, input)
}
