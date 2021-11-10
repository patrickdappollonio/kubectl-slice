package slice

import "testing"

func TestErrorsInterface(t *testing.T) {
	var _ error = &strictModeSkipErr{}
	var _ error = &skipErr{}
}
