package logger

import (
	"io"
	"testing"

	"github.com/stretchr/testify/require"
)

type testLogger struct {
	FnPrintln          func(...any)
	FnSetOutput        func(w io.Writer)
	FnPrintf           func(format string, v ...any)
	PrintlnCallCount   int
	SetOutputCallCount int
	PrintfCallCount    int
}

func (tl *testLogger) Println(v ...any) {
	tl.PrintlnCallCount++
	if tl.FnPrintln != nil {
		tl.FnPrintln(v...)
	}
}

func (tl *testLogger) SetOutput(w io.Writer) {
	tl.SetOutputCallCount++
	if tl.FnSetOutput != nil {
		tl.FnSetOutput(w)
	}
}

func (tl *testLogger) Printf(format string, v ...any) {
	tl.PrintfCallCount++
	if tl.FnPrintf != nil {
		tl.FnPrintf(format, v...)
	}
}

func TestLogger(t *testing.T) {
	var _ Logger = (*NOOP)(nil) // ensure NOOP satisfies Logger

	t.Run("call count", func(t *testing.T) {
		tl := &testLogger{}
		tl.Println("foo")
		tl.SetOutput(io.Discard)
		tl.Printf("foo", "bar")
		require.Equal(t, 1, tl.PrintlnCallCount)
		require.Equal(t, 1, tl.SetOutputCallCount)
		require.Equal(t, 1, tl.PrintfCallCount)
	})
}
