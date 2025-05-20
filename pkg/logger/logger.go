package logger

import "io"

// Logger is the interface used by Split to log debug messages
// and it's satisfied by Go's log.Logger
type Logger interface {
	Printf(format string, v ...any)
	SetOutput(w io.Writer)
	Println(v ...any)
}

// NOOP is a logger that does nothing
type NOOP struct{}

func (NOOP) Println(...any)            {}
func (NOOP) SetOutput(_ io.Writer)     {}
func (NOOP) Printf(_ string, _ ...any) {}

// NOOPLogger is a logger that does nothing
var NOOPLogger Logger = &NOOP{}
