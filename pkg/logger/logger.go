package logger

import "io"

// Logger is the interface used by Split to log debug messages
// and it's satisfied by Go's log.Logger
type Logger interface {
	Printf(format string, v ...any)
	SetOutput(w io.Writer)
	Println(v ...any)
}

// NOOP implements the Logger interface but performs no operations
type NOOP struct{}

func (NOOP) Println(...any)            {}
func (NOOP) SetOutput(_ io.Writer)     {}
func (NOOP) Printf(_ string, _ ...any) {}

// NOOPLogger is a pre-initialized no-operation logger instance that can be used for testing or when logging is disabled
var NOOPLogger Logger = &NOOP{}
