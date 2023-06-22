package workflow

// SimpleDAGWorkflowTransit defines the transit node of a directed acyclic graph.
type SimpleDAGWorkflowTransit struct {
	// The name of the transit node.
	//
	// Although it is not required to be unique, it is not recommended to duplicate the name or leave it blank,
	// and it is not recommended to use `input` or `output` because it may be a reserved name in the future.
	name string

	// channelInputs defines the input channel names required by this transit node.
	//
	// All channels are listened to when building a workflow.
	// The transit node will only be executed when all channels have passed in values.
	// Therefore, at least one channel must be specified.
	channelInputs []string

	// channelOutputs defines the output channel names required by this transit node.
	//
	// The results of the execution of the transit node will be sequentially sent to the designated output channel,
	// so that the subsequent node can continue to execute. Therefore, at least one channel must be specified.
	channelOutputs []string
	worker         func(...any) any
}

type SimpleDAGInterface[TInput, TOutput any] interface {
	InitChannels(channels ...string)
	AttachChannels(channels ...string)
	InitWorkflow(input string, output string, transits ...*SimpleDAGWorkflowTransit)
	AttachWorkflowTransit(...*SimpleDAGWorkflowTransit)
	BuildWorkflow()
	BuildWorkflowInput(result any, inputs ...string)
	BuildWorkflowOutput(outputs ...string) *[]any
	CloseWorkflow()
	Run(input *TInput) *TOutput
}
