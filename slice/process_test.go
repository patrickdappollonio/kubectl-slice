package slice

import (
	"testing"
)

func Test_inSliceIgnoreCase(t *testing.T) {
	type args struct {
		slice    []string
		expected string
	}

	tests := []struct {
		name string
		args args
		want bool
	}{
		{
			name: "in slice",
			args: args{
				slice:    []string{"foo", "bar"},
				expected: "foo",
			},
			want: true,
		},
		{
			name: "not in slice",
			args: args{
				slice:    []string{"foo", "bar"},
				expected: "baz",
			},
			want: false,
		},
		{
			name: "pattern fo-star without glob support",
			args: args{
				slice:    []string{"fo*", "bar"},
				expected: "foo",
			},
			want: false,
		},
		{
			name: "pattern anything without glob support",
			args: args{
				slice:    []string{"*"},
				expected: "foo",
			},
			want: false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if got := inSliceIgnoreCase(tt.args.slice, tt.args.expected); got != tt.want {
				t.Errorf("inSliceIgnoreCase() = %v, want %v", got, tt.want)
			}
		})
	}
}

func Test_inSliceIgnoreCaseGlob(t *testing.T) {
	type args struct {
		slice    []string
		expected string
	}

	tests := []struct {
		name string
		args args
		want bool
	}{
		{
			name: "in slice",
			args: args{
				slice:    []string{"foo", "bar"},
				expected: "foo",
			},
			want: true,
		},
		{
			name: "not in slice",
			args: args{
				slice:    []string{"foo", "bar"},
				expected: "baz",
			},
			want: false,
		},
		{
			name: "pattern fo-star",
			args: args{
				slice:    []string{"fo*", "bar"},
				expected: "foo",
			},
			want: true,
		},
		{
			name: "pattern anything",
			args: args{
				slice:    []string{"*"},
				expected: "foo",
			},
			want: true,
		},
		{
			name: "kubernetes annotation",
			args: args{
				slice:    []string{"kubernetes.io/*"},
				expected: "kubernetes.io/ingress.class",
			},
			want: true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if got := inSliceIgnoreCaseGlob(tt.args.slice, tt.args.expected); got != tt.want {
				t.Errorf("inSliceIgnoreCase() = %v, want %v", got, tt.want)
			}
		})
	}
}

func Test_checkStringInMap(t *testing.T) {
	type args struct {
		local map[string]interface{}
		key   string
	}

	tests := []struct {
		name string
		args args
		want string
	}{
		{
			name: "key found in map",
			args: args{
				local: map[string]interface{}{
					"foo": "bar",
				},
				key: "foo",
			},
			want: "bar",
		},
		{
			name: "key not found in map",
			args: args{
				local: map[string]interface{}{
					"foo": "bar",
				},
				key: "baz",
			},
			want: "",
		},
		{
			name: "key with non string value",
			args: args{
				local: map[string]interface{}{
					"foo": map[string]string{
						"bar": "baz",
					},
				},
				key: "foo",
			},
			want: "",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if str := checkStringInMap(tt.args.local, tt.args.key); str != tt.want {
				t.Errorf("checkStringInMap() = %v, want %v", str, tt.want)
			}
		})
	}
}
