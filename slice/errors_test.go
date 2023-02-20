package slice

import (
	"testing"

	"github.com/stretchr/testify/require"
)

func TestErrorsInterface(t *testing.T) {
	require.Implementsf(t, (*error)(nil), &strictModeSkipErr{}, "strictModeSkipErr should implement error")
	require.Implementsf(t, (*error)(nil), &skipErr{}, "skipErr should implement error")
}

func requireErrorIf(t *testing.T, wantErr bool, err error) {
	if wantErr {
		require.Error(t, err)
	} else {
		require.NoError(t, err)
	}
}
