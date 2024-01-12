// Copyright (c) 2023 - 2024 vistart. All rights reserved.
// Use of this source code is governed by Apache-2.0 license
// that can be found in the LICENSE file.

package simple

import (
	"context"
	"fmt"
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
	Transit() *Transit
	Level() LogLevel
	Message() string
}

type LogEventErrValueType struct {
	LogEventInterface
	err ErrValueType
}

func (l *LogEventErrValueType) Message() string {
	return l.err.Error()
}

func (l *LogEventErrValueType) Level() LogLevel {
	return LevelError
}

func (l *LogEventErrValueType) Transit() *Transit {
	return nil
}

type LogEventTransitReportedError struct {
	LogEventInterface
	err     error
	transit *Transit
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

type LogEventTransitStart struct {
	LogEventInterface
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

type LogEventTransitEnd struct {
	LogEventInterface
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

type LogEventTransitCanceled struct {
	LogEventInterface
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

type LogEventTransitWorkerPanicked struct {
	LogEventInterface
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

// LoggerInterface defines the logging method and the parameters required by the logger.
// For specific usage, please refer to Logger.
type LoggerInterface interface {
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
	transit := event.Transit()
	name := "<nil>"
	if transit != nil {
		name = transit.name
	}
	fmt.Printf("[GO-DAG] %v |%s %10s %s| %s\n", time.Now().Format(l.params.TimestampFormat), color, name, reset, event.Message())
}

func (l *Logger) Log(ctx context.Context, events ...LogEventInterface) {
	for _, event := range events {
		l.logEvent(ctx, event)
	}
}
