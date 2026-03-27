package logger

import (
	"context"
	"time"
	"github.com/wailsapp/wails/v2/pkg/runtime"
)

type Level string

const (
	INFO  Level = "info"
	WARN  Level = "warn"
	ERROR Level = "error"
)

type Emitter interface {
	Emit(level Level, msg string)
}

type LogEmitter struct {
	ctx context.Context
}

func NewLogEmitter(ctx context.Context) *LogEmitter {
	return &LogEmitter{ctx: ctx}
}

type LogEvent struct {
	Level     Level             `json:"level"`
	Msg       string            `json:"message"`
	Source    string            `json:"source"`
	Timestamp string            `json:"timestamp"`
	Fields    map[string]string `json:"fields,omitempty"`
}

func (e *LogEmitter) emit(level Level, msg string, source string, fields map[string]string) {
	runtime.EventsEmit(e.ctx, "log:", LogEvent{
		Level:     level,
		Msg:       msg,
		Source:    source,
		Timestamp: time.Now().Format("2006-01-02 15:04:05"),
		Fields:    fields,
	})
}
func (e *LogEmitter) Info(msg, source string, fields map[string]string) {
	e.emit(INFO, msg, source, fields)
}

func (e *LogEmitter) Warn(msg, source string, fields map[string]string) {
	e.emit(WARN, msg, source, fields)
}

func (e *LogEmitter) Error(msg, source string, fields map[string]string) {
	e.emit(ERROR, msg, source, fields)
}
