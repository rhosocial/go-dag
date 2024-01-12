// Copyright (c) 2023 - 2024 vistart. All rights reserved.
// Use of this source code is governed by Apache-2.0 license
// that can be found in the LICENSE file.

package simple

import (
	"context"
	"fmt"
	"sync"
	"time"
)

type LogLevel int

const (
	LevelDebug LogLevel = iota
	LevelInfo
	LevelWarning
	LevelError
)

type LogEventInterface interface {
	Name() string
	Level() LogLevel
	Message() string
}

type LogEventErrorInterface interface {
	Error() error
}

type LogEventTransitInterface interface {
	Transit() *Transit
}

type LogEventError struct {
	LogEventErrorInterface
	err error
}

func NewLogEventError(err error) *LogEventError {
	return &LogEventError{err: err}
}

func (l *LogEventError) Error() error {
	return l.err
}

type LogEventErrorValueTypeMismatch struct {
	LogEventInterface
	LogEventError
	err ErrValueTypeMismatch
}

func (l *LogEventErrorValueTypeMismatch) Message() string {
	return l.err.Error()
}

func (l *LogEventErrorValueTypeMismatch) Level() LogLevel {
	return LevelError
}

func (l *LogEventErrorValueTypeMismatch) Name() string {
	return "final chn"
}

type LogEventTransitReportedError struct {
	LogEventInterface
	LogEventTransitInterface
	LogEventError
	transit *Transit
}

func NewLogEventTransitReportedError(transit *Transit, err error) *LogEventTransitReportedError {
	return &LogEventTransitReportedError{transit: transit, LogEventError: *NewLogEventError(err)}
}

func (l *LogEventTransitReportedError) Message() string {
	return l.err.Error()
}

func (l *LogEventTransitReportedError) Level() LogLevel {
	return LevelWarning
}

func (l *LogEventTransitReportedError) Transit() *Transit {
	return l.transit
}

func (l *LogEventTransitReportedError) Name() string {
	if l.transit == nil {
		return "<nil>"
	}
	return l.transit.name
}

type LogEventTransitStart struct {
	LogEventInterface
	LogEventTransitInterface
	transit *Transit
}

func (l *LogEventTransitStart) Message() string {
	return "starting..."
}

func (l *LogEventTransitStart) Transit() *Transit {
	return l.transit
}

func (l *LogEventTransitStart) Level() LogLevel {
	return LevelDebug
}

func (l *LogEventTransitStart) Name() string {
	if l.transit == nil {
		return "<nil>"
	}
	return l.transit.name
}

type LogEventTransitEnd struct {
	LogEventInterface
	LogEventTransitInterface
	transit *Transit
}

func (l *LogEventTransitEnd) Message() string {
	return "ended."
}

func (l *LogEventTransitEnd) Transit() *Transit {
	return l.transit
}

func (l *LogEventTransitEnd) Level() LogLevel {
	return LevelDebug
}

func (l *LogEventTransitEnd) Name() string {
	if l.transit == nil {
		return "<nil>"
	}
	return l.transit.name
}

type LogEventTransitCanceled struct {
	LogEventInterface
	LogEventTransitInterface
	LogEventError
	transit *Transit
}

func (l *LogEventTransitCanceled) Message() string {
	return "cancellation notified."
}

func (l *LogEventTransitCanceled) Transit() *Transit {
	return l.transit
}

func (l *LogEventTransitCanceled) Level() LogLevel {
	return LevelWarning
}

func (l *LogEventTransitCanceled) Name() string {
	if l.transit == nil {
		return "<nil>"
	}
	return l.transit.name
}

type LogEventTransitWorkerPanicked struct {
	LogEventInterface
	LogEventTransitInterface
	LogEventError
	transit *Transit
	err     ErrWorkerPanicked
}

func (l *LogEventTransitWorkerPanicked) Message() string {
	return "worker panicked."
}

func (l *LogEventTransitWorkerPanicked) Transit() *Transit {
	return l.transit
}

func (l *LogEventTransitWorkerPanicked) Level() LogLevel {
	return LevelError
}
func (l *LogEventTransitWorkerPanicked) Name() string {
	if l.transit == nil {
		return "<nil>"
	}
	return l.transit.name
}

// LoggerInterface defines the logging method and the parameters required by the logger.
// For specific usage, please refer to Logger.
type LoggerInterface interface {
	// Log an event.
	Log(ctx context.Context, events ...LogEventInterface)

	SetFlags(uint)
}

type LoggerParams struct {
	TimestampFormat string
	Caller          bool
	logDebugEnabled bool
}

type Logger struct {
	params LoggerParams
	LoggerInterface
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

const (
	LDebugEnabled = 2
)

func (l *Logger) SetFlags(flags uint) {
	l.params.logDebugEnabled = flags&LDebugEnabled > 0
}

func (l *Logger) logEvent(ctx context.Context, event LogEventInterface) {
	if !l.params.logDebugEnabled && (event.Level() == LevelDebug) {
		return
	}
	color := green
	if event.Level() == LevelWarning {
		color = yellow
	} else if event.Level() == LevelError {
		color = red
	}
	fmt.Printf("[GO-DAG] %v |%s %10s %s| %s\n", time.Now().Format(l.params.TimestampFormat), color, event.Name(), reset, event.Message())
}

func (l *Logger) Log(ctx context.Context, events ...LogEventInterface) {
	for _, event := range events {
		l.logEvent(ctx, event)
	}
}

// ErrorCollectorInterface defines the methods that error collectors should implement.
// It is also a logger, so it also needs to implement all methods specified by the LoggerInterface.
type ErrorCollectorInterface interface {
	LoggerInterface
	Listen(ctx context.Context)
	Get() []LogEventErrorInterface

	Append(event *LogEventErrorInterface)
}

type ErrorCollector struct {
	ErrorCollectorInterface
	mu       sync.RWMutex
	errors   []LogEventErrorInterface
	listener chan LogEventErrorInterface
}

func NewErrorCollector() *ErrorCollector {
	return &ErrorCollector{
		errors:   make([]LogEventErrorInterface, 0),
		listener: make(chan LogEventErrorInterface),
	}
}

func (l *ErrorCollector) Listen(ctx context.Context) {
	var e LogEventErrorInterface
	for {
		select {
		case <-ctx.Done():
			return
		case e = <-l.listener:
			l.Append(&e)
		default:
		}
	}
}

func (l *ErrorCollector) Get() []LogEventErrorInterface {
	l.mu.RLock()
	defer l.mu.RUnlock()
	return l.errors
}

func (l *ErrorCollector) Append(event *LogEventErrorInterface) {
	l.mu.Lock()
	defer l.mu.Unlock()
	l.errors = append(l.errors, *event)
}

func (l *ErrorCollector) Log(ctx context.Context, events ...LogEventInterface) {
	for _, event := range events {
		if e, ok := event.(LogEventErrorInterface); ok && event != nil {
			l.listener <- e
		}
	}
}

func (l *ErrorCollector) SetFlags(uint) {}
