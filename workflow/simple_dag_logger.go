package workflow

import (
	"encoding/json"
	"fmt"
	"runtime"
	"time"
)

// SimpleDAGLogger defines the logging method and the parameters required by the logger.
// For specific usage, please refer to SimpleDAGJSONLogger.
type SimpleDAGLogger interface {
	Log(level LogLevel, message string, args ...any)
	Trace(level LogLevel, transit *SimpleDAGWorkflowTransit, message string, args ...any)

	SetParams(params SimpleDAGLogParams)
	GetParams() SimpleDAGLogParams
}

type LogLevel int

const (
	LevelDebug LogLevel = iota
	LevelInfo
	LevelWarning
	LevelError
)

type SimpleDAGLogParams struct {
	TimestampFormat string
	Caller          bool
	logDebugEnabled bool
	ExtraParams     map[string]any
}

type SimpleDAGJSONLogger struct {
	params SimpleDAGLogParams
}

func NewSimpleDAGJSONLogger() *SimpleDAGJSONLogger {
	return &SimpleDAGJSONLogger{
		params: SimpleDAGLogParams{
			TimestampFormat: "2006-01-02 15:04:05.000000",
			Caller:          true,
		},
	}
}

const (
	LDebugEnabled = 2
)

func (l *SimpleDAGJSONLogger) SetFlags(flags uint) {
	l.params.logDebugEnabled = flags&LDebugEnabled > 0
}

func (l *SimpleDAGJSONLogger) Log(level LogLevel, message string, args ...any) {
	if !l.params.logDebugEnabled && (level == LevelDebug) {
		return
	}
	data := map[string]any{
		"timestamp": time.Now().Format(l.params.TimestampFormat),
		"message":   message,
	}

	if l.params.Caller {
		if pc, _, _, ok := runtime.Caller(1); ok {
			fn := runtime.FuncForPC(pc)
			data["caller"] = fn.Name()
		}
	}

	if b, err := json.Marshal(data); err != nil {
		panic(err)
	} else {
		fmt.Print(string(b))
	}
}

func (l *SimpleDAGJSONLogger) SetParams(params SimpleDAGLogParams) {
	l.params = params
}

func (l *SimpleDAGJSONLogger) GetParams() SimpleDAGLogParams {
	return l.params
}

const (
	green   = "\033[97;42m"
	white   = "\033[90;47m"
	yellow  = "\033[90;43m"
	red     = "\033[97;41m"
	blue    = "\033[97;44m"
	magenta = "\033[97;45m"
	cyan    = "\033[97;46m"
	reset   = "\033[0m"
)

// Trace output trace information.
// level refers to the log information level.
// transit refers to the tracking transit.
// message refers to the tracking message.
// args refers to other parameters.
// By default, LevelDebug logs are not displayed. If you want to display, call SetFlags(LDebugEnabled)
func (l *SimpleDAGJSONLogger) Trace(level LogLevel, transit *SimpleDAGWorkflowTransit, message string, args ...any) {
	if !l.params.logDebugEnabled && (level == LevelDebug) {
		return
	}
	color := green
	if level == LevelWarning {
		color = yellow
	} else if level == LevelError {
		color = red
	}
	fmt.Printf("[GO-DAG] %v |%s %10s %s| %s\n", time.Now().Format(l.params.TimestampFormat), color, transit.name, reset, message)
}
