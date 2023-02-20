package slice

import (
	"testing"

	"github.com/stretchr/testify/require"
)

func TestErrorsInterface(t *testing.T) {
	require.Implementsf(t, (*error)(nil), &strictModeSkipErr{}, "strictModeSkipErr should implement error")
	require.Implementsf(t, (*error)(nil), &skipErr{}, "skipErr should implement error")
}
