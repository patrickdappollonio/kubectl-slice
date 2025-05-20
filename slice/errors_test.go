package slice

import (
	"testing"

	"github.com/patrickdappollonio/kubectl-slice/pkg/errors"
	"github.com/stretchr/testify/require"
)

func TestErrorsInterface(t *testing.T) {
	require.Implementsf(t, (*error)(nil), &errors.StrictModeSkipErr{}, "StrictModeSkipErr should implement error")
	require.Implementsf(t, (*error)(nil), &errors.SkipErr{}, "SkipErr should implement error")
	require.Implementsf(t, (*error)(nil), &errors.CantFindFieldErr{}, "CantFindFieldErr should implement error")
}

func requireErrorIf(t *testing.T, wantErr bool, err error) {
	if wantErr {
		require.Error(t, err)
	} else {
		require.NoError(t, err)
	}
}
