package workflow

import "context"

type StandardDAGInterface[TInput, TOutput any] interface {
	SimpleDAGInitInterface
	BuildWorkflow(ctx context.Context) error
	BuildWorkflowInput(result any, inputs ...string)
	BuildWorkflowOutput(ctx context.Context, outputs ...string) *[]any
	CloseWorkflow()
	Execute(ctx context.Context, input *TInput) *TOutput
	RunOnce(ctx context.Context, input *TInput) *TOutput
}
