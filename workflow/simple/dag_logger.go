// Copyright (c) 2023 - 2024 vistart. All rights reserved.
// Use of this source code is governed by Apache-2.0 license
// that can be found in the LICENSE file.

package simple

import (
	"context"
	"encoding/json"
	"fmt"
	"runtime"
	"time"
)

// LoggerInterface defines the logging method and the parameters required by the logger.
// For specific usage, please refer to Logger.
type LoggerInterface interface {
	Log(ctx context.Context, level LogLevel, message string, args ...any)
	Trace(ctx context.Context, level LogLevel, transit *Transit, message string, args ...any)

	SetFlags(uint)
}

type LogLevel int

const (
	LevelDebug LogLevel = iota
	LevelInfo
	LevelWarning
	LevelError
)

type LoggerParams struct {
	TimestampFormat string
	Caller          bool
	logDebugEnabled bool
	ExtraParams     map[string]any
}

type Logger struct {
	params LoggerParams
	LoggerInterface
}

const (
	LDebugEnabled = 2
)

func (l *Logger) SetFlags(flags uint) {
	l.params.logDebugEnabled = flags&LDebugEnabled > 0
}

func (l *Logger) Log(ctx context.Context, level LogLevel, message string, args ...any) {
	if !l.params.logDebugEnabled && (level == LevelDebug) {
		return
	}
	data := map[string]any{
		"timestamp": time.Now().Format(l.params.TimestampFormat),
		"message":   message,
	}

	if len(args) > 0 {
		data["args"] = args
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
func (l *Logger) Trace(ctx context.Context, level LogLevel, transit *Transit, message string, args ...any) {
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
