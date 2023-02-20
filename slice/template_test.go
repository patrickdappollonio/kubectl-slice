package slice

import (
	"log"
	"testing"
)

type fields struct {
	opts Options
	log  *log.Logger
}

func newFields(opts Options) fields {
	return fields{
		opts: opts,
		log:  log.Default(),
	}
}

func TestSplit_compileTemplate(t *testing.T) {
	tests := []struct {
		name    string
		fields  fields
		wantErr bool
	}{
		{
			name: "compile template generic",
			fields: newFields(Options{
				GoTemplate: "{{.}}",
			}),
		},
		{
			name: "non existent function",
			fields: newFields(Options{
				GoTemplate: "{{. | foobarbaz}}",
			}),
			wantErr: true,
		},
		{
			name: "existent function",
			fields: newFields(Options{
				GoTemplate: "{{. | lower}}",
			}),
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			s := &Split{
				opts: tt.fields.opts,
				log:  tt.fields.log,
			}
			if err := s.compileTemplate(); (err != nil) != tt.wantErr {
				t.Errorf("compile template error: recv = %v, wanted %v", err, tt.wantErr)
			}
		})
	}
}
