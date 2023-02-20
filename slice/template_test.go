package slice

import (
	"testing"

	"github.com/stretchr/testify/require"
)

func TestTemplate_compileTemplate(t *testing.T) {
	tests := []struct {
		name    string
		opts    Options
		wantErr bool
	}{
		{
			name: "compile template generic",
			opts: Options{
				GoTemplate: "{{.}}",
			},
		},
		{
			name: "non existent function",
			opts: Options{
				GoTemplate: "{{. | foobarbaz}}",
			},
			wantErr: true,
		},
		{
			name: "existent function",
			opts: Options{
				GoTemplate: "{{. | lower}}",
			},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			s := &Split{opts: tt.opts, log: nolog}

			err := s.compileTemplate()

			if tt.wantErr {
				require.Error(t, err)
			} else {
				require.NoError(t, err)
			}
		})
	}
}
