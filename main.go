package main

import (
	"context"
	"fmt"
	"log"
	"reflect"
	"strings"
	"sync"
	"time"
)

func main() {
}

func subtest(ctx context.Context, name string, doneCallback func()) {
	index := 0
	for {
		time.Sleep(100 * time.Millisecond)
		index++
		select {
		case <-ctx.Done():
			log.Printf("%s: %s\n", name, context.Cause(ctx))
			if value, ok := ctx.Value("parent").(string); ok {
				log.Printf("%s: parent:%s\n", name, value)
			}
			if doneCallback != nil {
				doneCallback()
			}
			return
		default:
			log.Printf("%s: %d working...\n", name, index)
		}
	}
}

func subtesterr(ctx context.Context, name string, doneCallback func()) error {
	index := 0
	for {
		time.Sleep(100 * time.Millisecond)
		index++
		select {
		case <-ctx.Done():
			log.Printf("%s: %s\n", name, context.Cause(ctx))
			if value, ok := ctx.Value("parent").(string); ok {
				log.Printf("%s: parent:%s\n", name, value)
			}
			if doneCallback != nil {
				doneCallback()
			}
			return ctx.Err()
		default:
			log.Printf("%s: %d working...\n", name, index)
		}
	}
	return nil
}

// clientChan must be initialized before use
var clientChan chan string

func producer(content string) {
	clientChan <- content
}

func consumer() *string {
	if content, ok := <-clientChan; ok {
		return &content
	}
	return nil
}

// dagChanMap must be initialized before use
var dagChanMap map[string]chan string

func dagInput(content string, chanNames ...string) {
	for _, next := range chanNames {
		next := next
		go func(next string) {
			dagChanMap[next] <- content
		}(next)
	}
}

func dagOutput(chanNames ...string) *string {
	var count = len(chanNames)
	var contents = make([]string, count)
	var wg sync.WaitGroup
	wg.Add(count)
	for i, name := range chanNames {
		i := i
		name := name
		go func(i int, name string) {
			contents[i] = <-dagChanMap[name]
			wg.Done()
		}(i, name)
	}
	wg.Wait()
	var content = strings.Join(contents, ",")
	return &content
}

type DAGWorkflowTransit struct {
	Name           string
	channelInputs  []string
	channelOutputs []string
	worker         func(...any) any
}

type DAGInterface[TInput, TOutput any] interface {
	InitChannels(channels ...string)
	AttachChannels(channels ...string)
	InitWorkflow(input string, output string, transits ...*DAGWorkflowTransit)
	AttachWorkflowTransit(...*DAGWorkflowTransit)
	BuildWorkflow()
	BuildWorkflowInput(result any, inputs ...string)
	BuildWorkflowOutput(outputs ...string) *[]any
	CloseWorkflow()
	Run(input *TInput) *TOutput
}

type DAG[TInput, TOutput any] struct {
	channels         map[string]chan any
	workflowInput    string
	workflowOutput   string
	workflowTransits []*DAGWorkflowTransit
	DAGInterface[TInput, TOutput]
}

// InitChannels initializes channels for workflows.
//
// The parameter is the channel name. The channel name cannot be repeated, otherwise the last input shall prevail.
func (d *DAG[TInput, TOutput]) InitChannels(channels ...string) {
	if channels == nil || len(channels) == 0 {
		return
	}
	d.channels = make(map[string]chan any, len(channels))
	for _, v := range channels {
		v := v
		d.channels[v] = make(chan any)
	}
}

func (d *DAG[TInput, TOutput]) AttachChannels(channels ...string) {
	if channels == nil || len(channels) == 0 {
		return
	}
	if d.channels == nil || len(d.channels) == 0 {
		d.InitChannels(channels...)
		return
	}
	for _, v := range channels {
		v := v
		d.channels[v] = make(chan any)
	}
}

// InitWorkflow initializes workflows.
//
// The parameter is the name of the workflow node.
func (d *DAG[TInput, TOutput]) InitWorkflow(input string, output string, transits ...*DAGWorkflowTransit) {
	d.workflowInput = input
	d.workflowOutput = output
	lenTransits := len(transits)
	if lenTransits == 0 {
		return
	}
	d.workflowTransits = make([]*DAGWorkflowTransit, lenTransits)
	for i, t := range transits {
		d.workflowTransits[i] = t
	}
}

func (d *DAG[TInput, TOutput]) AttachWorkflowTransit(transits ...*DAGWorkflowTransit) {
	for _, t := range transits {
		d.workflowTransits = append(d.workflowTransits, t)
	}
}

func (d *DAG[TInput, TOutput]) BuildWorkflowInput(result any, inputs ...string) {
	for _, next := range inputs {
		next := next
		if _, existed := d.channels[next]; !existed {
			panic(fmt.Sprintf("Specified channel[%s] does not exist, not initialized?", next))
		}
		go func(next string) {
			log.Println(fmt.Sprintf("WorkflowInput[channel: %s] preparing: ", next), result)
			d.channels[next] <- result
			log.Println(fmt.Sprintf("WorkflowInput[channel: %s] sent: ", next), result)
		}(next)
	}
}

func (d *DAG[TInput, TOutput]) BuildWorkflowOutput(outputs ...string) *[]any {
	var count = len(outputs)
	log.Println("WorkflowOutput will receive the contents from: ", outputs)
	var results = make([]any, count)
	var wg sync.WaitGroup
	wg.Add(count)
	for i, name := range outputs {
		//i := i
		//name := name
		go func(i int, name string) {
			defer wg.Done()
			if _, existed := d.channels[name]; !existed {
				panic(fmt.Sprintf("Specified channel[%s] does not exist, not initialized?", name))
			}
			log.Println(fmt.Sprintf("WorkflowOutput[channel: %s] listening...", name))
			results[i] = <-d.channels[name]
			log.Println(fmt.Sprintf("WorkflowOutput[channel: %s] received: ", name), results[i])
		}(i, name)
	}
	wg.Wait()
	log.Println(fmt.Sprintf("WorkflowOutput len: %d", len(results)))
	return &results
}

func (d *DAG[TInput, TOutput]) BuildWorkflow() {
	// Build Input
	if len(d.workflowInput) == 0 {
		return
	}

	// Build Output
	if len(d.workflowOutput) == 0 {
		return
	}

	// Build Transits
	if len(d.workflowTransits) == 0 {
		return
	}
	for _, t := range d.workflowTransits {
		go func(t *DAGWorkflowTransit) {
			var results = d.BuildWorkflowOutput(t.channelInputs...)

			var result = t.worker(*results...)

			d.BuildWorkflowInput(result, t.channelOutputs...)
		}(t)
	}
}

func (d *DAG[TInput, TOutput]) CloseWorkflow() {
	if len(d.workflowInput) == 0 {
		return
	}
	//close(d.channels[d.workflowInput])
	if len(d.workflowOutput) == 0 {
		return
	}
	//close(d.channels[d.workflowOutput])
	if len(d.workflowTransits) == 0 {
		return
	}
	for _, t := range d.workflowTransits {
		for _, c := range t.channelOutputs {
			close(d.channels[c])
		}
	}
}

func (d *DAG[TInput, TOutput]) Run(input *TInput) *TOutput {
	d.BuildWorkflow()
	defer d.CloseWorkflow()

	var results TOutput
	var wg sync.WaitGroup
	wg.Add(1)
	go func() {
		defer wg.Done()
		r := d.BuildWorkflowOutput(d.workflowOutput)
		log.Println("final output:", (*r)[0])
		if r, ok := (*r)[0].(TOutput); !ok {
			t := new(TOutput)
			panic(fmt.Sprintf("The type of the value received[%s] is inconsistent with the target[%s].", reflect.TypeOf(r), reflect.TypeOf(t)))
		} else {
			results = r
		}
	}()
	d.BuildWorkflowInput(*input, d.workflowInput)
	defer close(d.channels[d.workflowInput])
	log.Println("input sent:", *input)
	wg.Wait()
	log.Println("output received:", results)
	return &results
}
