package slice

import "testing"

func TestErrorsInterface(t *testing.T) {
	var _ error = &strictModeErr{}
	var _ error = &kindSkipErr{}
}
