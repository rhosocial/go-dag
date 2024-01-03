package workflow

import (
	"errors"
	"fmt"
	"reflect"
	"strings"
)

var ErrChannelNotInitialized = errors.New("the channel map is not initialized")
var ErrChannelNotExist = errors.New("the specified channel does not exist")
var ErrChannelInputEmpty = errors.New("the input channel is empty")
var ErrChannelOutputEmpty = errors.New("the output channel is empty")

type ErrDAGChannelNameExisted struct {
	name string
	error
}

func (e *ErrDAGChannelNameExisted) Error() string {
	return fmt.Sprintf("the channel[%s] has existed.", e.name)
}

// ErrSimpleDAGValueType defines that the data type output by the node is inconsistent with expectation.
type ErrSimpleDAGValueType struct {
	expect any
	actual any
	error
}

func (e ErrSimpleDAGValueType) Error() string {
	return fmt.Sprintf("The type of the value [%s] is inconsistent with expectation [%s].",
		reflect.TypeOf(e.actual), reflect.TypeOf(e.expect))
}

// ErrRedundantChannels indicates that there are unused channelInputs.
type ErrRedundantChannels struct {
	channels []string
	error
}

func (e ErrRedundantChannels) Error() string {
	return fmt.Sprintf("Redundant channelInputs: %v", strings.Join(e.channels, ", "))
}

// ErrTransitChannelNonExist indicates that the channel(s) to be used by the specified node does not exist.
type ErrTransitChannelNonExist struct {
	transitName    string
	channelInputs  []string
	channelOutputs []string
	error
}

func (e ErrTransitChannelNonExist) Error() string {
	return fmt.Sprintf("The specified channel(s) does not exist: input[%v], output[%v]",
		strings.Join(e.channelInputs, ", "), strings.Join(e.channelOutputs, ", "))
}
