package slice

import (
	"testing"
)

func TestSplit_validateFilters(t *testing.T) {
	tests := []struct {
		name    string
		opts    Options
		wantErr bool
	}{
		{
			name: "prevent using allow skipping kind while using included kinds",
			opts: Options{
				AllowEmptyKinds: true,
				IncludedKinds:   []string{"foo"},
			},
			wantErr: true,
		},
		{
			name: "prevent using allow skipping kind while using excluded kinds",
			opts: Options{
				AllowEmptyKinds: true,
				ExcludedKinds:   []string{"foo"},
			},
			wantErr: true,
		},
		{
			name: "prevent using allow skipping name while using included names",
			opts: Options{
				AllowEmptyNames: true,
				IncludedNames:   []string{"foo"},
			},
			wantErr: true,
		},
		{
			name: "prevent using allow skipping name while using excluded names",
			opts: Options{
				AllowEmptyNames: true,
				ExcludedNames:   []string{"foo"},
			},
			wantErr: true,
		},
		{
			name: "cannot specify included and excluded kinds",
			opts: Options{
				IncludedKinds: []string{"foo"},
				ExcludedKinds: []string{"bar"},
			},
			wantErr: true,
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			s := &Split{opts: tt.opts}
			if err := s.validateFilters(); (err != nil) != tt.wantErr {
				t.Errorf("Split.validateFilters() error = %v, wantErr %v", err, tt.wantErr)
			}
		})
	}
}
