package template

import (
	"math/rand"
	"os"
	"strings"
	"testing"
	"time"

	"github.com/stretchr/testify/require"
)

func Test_mapValueByIndex(t *testing.T) {
	tests := []struct {
		name    string
		index   string
		m       map[string]interface{}
		want    interface{}
		wantErr bool
	}{
		{
			name:    "nil map",
			index:   "foo",
			m:       nil,
			wantErr: true,
		},
		{
			name:    "empty index",
			index:   "",
			m:       map[string]interface{}{},
			wantErr: true,
		},
		{
			name:  "fetch existent field",
			index: "foo",
			m: map[string]interface{}{
				"foo": "bar",
			},
			want: "bar",
		},
		{
			name:  "fetch nonexistent field",
			index: "baz",
			m: map[string]interface{}{
				"foo": "bar",
			},
			wantErr: true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got, err := mapValueByIndex(tt.index, tt.m)
			requireErrorIf(t, tt.wantErr, err)
			require.Equal(t, tt.want, got)
		})
	}
}

func Test_strJSON(t *testing.T) {
	tests := []struct {
		name    string
		val     interface{}
		want    string
		wantErr bool
	}{
		{
			name: "string conversion",
			val:  "foo",
			want: "foo",
		},
		{
			name: "bool true conversion",
			val:  true,
			want: "true",
		},
		{
			name: "bool false conversion",
			val:  false,
			want: "false",
		},
		{
			name: "float64 conversion",
			val:  3.141592654,
			want: "3.141592654",
		},
		{
			name:    "incorrect data type conversion",
			val:     []string{},
			wantErr: true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got, err := strJSON(tt.val)
			requireErrorIf(t, tt.wantErr, err)
			require.Equal(t, tt.want, got)
		})
	}
}

func Test_jsonAlphanumify(t *testing.T) {
	tests := []struct {
		name    string
		val     interface{}
		want    string
		wantErr bool
	}{
		{
			name: "remove dots",
			val:  "foo.bar",
			want: "foobar",
		},
		{
			name: "remove dots and slashes",
			val:  "foo.bar/baz",
			want: "foobarbaz",
		},
		{
			name: "remove all special characters",
			val:  "foo.bar/baz!@#$%^&*()_+-=[]{}\\|;:'\",<.>/?",
			want: "foobarbaz",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got, err := jsonAlphanumify(tt.val)
			requireErrorIf(t, tt.wantErr, err)
			require.Equal(t, tt.want, got)
		})
	}
}

func Test_jsonAlphanumdash(t *testing.T) {
	tests := []struct {
		name    string
		val     interface{}
		want    string
		wantErr bool
	}{
		{
			name: "remove dots",
			val:  "foo.bar-baz",
			want: "foobar-baz",
		},
		{
			name: "remove dots and slashes",
			val:  "foo.bar/baz-daz",
			want: "foobarbaz-daz",
		},
		{
			name: "remove all special characters",
			val:  "foo.bar/baz!@#$%^&*()_+=[]{}\\|;:'\",<.>/?-daz",
			want: "foobarbaz-daz",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got, err := jsonAlphanumdash(tt.val)
			requireErrorIf(t, tt.wantErr, err)
			require.Equal(t, tt.want, got)
		})
	}
}

func Test_jsonDotToDash(t *testing.T) {
	tests := []struct {
		name    string
		val     interface{}
		want    string
		wantErr bool
	}{
		{
			name: "single dot",
			val:  "foo.bar",
			want: "foo-bar",
		},
		{
			name: "multi dot",
			val:  "foo...bar",
			want: "foo---bar",
		},
		{
			name: "no dot",
			val:  "foobar",
			want: "foobar",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got, err := jsonDotToDash(tt.val)
			requireErrorIf(t, tt.wantErr, err)
			require.Equal(t, tt.want, got)
		})
	}
}

func Test_jsonDotToUnder(t *testing.T) {
	tests := []struct {
		name    string
		val     interface{}
		want    string
		wantErr bool
	}{
		{
			name: "single dot",
			val:  "foo.bar",
			want: "foo_bar",
		},
		{
			name: "multi dot",
			val:  "foo...bar",
			want: "foo___bar",
		},
		{
			name: "no dot",
			val:  "foobar",
			want: "foobar",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got, err := jsonDotToUnder(tt.val)
			requireErrorIf(t, tt.wantErr, err)
			require.Equal(t, tt.want, got)
		})
	}
}

func Test_jsonReplace(t *testing.T) {
	type args struct {
		search  string
		replace string
		val     interface{}
	}

	tests := []struct {
		name    string
		args    args
		want    string
		wantErr bool
	}{
		{
			name: "basic replace",
			args: args{
				search:  "foo",
				replace: "bar",
				val:     "foobar",
			},
			want: "barbar",
		},
		{
			name: "non existent replacement",
			args: args{
				search:  "foo",
				replace: "bar",
				val:     "barbar",
			},
			want: "barbar",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got, err := jsonReplace(tt.args.search, tt.args.replace, tt.args.val)
			requireErrorIf(t, tt.wantErr, err)
			require.Equal(t, tt.want, got)
		})
	}
}

func Test_env(t *testing.T) {
	letters := []rune("abcdefghijklmnopqrstuvwxyz")

	randSeq := func(n int) string {
		rnd := rand.New(rand.NewSource(time.Now().UnixNano()))
		b := make([]rune, n)
		for i := range b {
			b[i] = letters[rnd.Intn(len(letters))]
		}
		return string(b)
	}

	type args struct {
		key string
		env map[string]string
	}

	tests := []struct {
		name string
		args args
		want string
	}{
		{
			name: "generic",
			args: args{
				key: "foo",
				env: map[string]string{
					"foo": "bar",
				},
			},
			want: "bar",
		},
		{
			name: "non-existent",
			args: args{
				key: "fooofooo",
				env: map[string]string{
					"foo": "bar",
				},
			},
			want: "",
		},
		{
			name: "case insensitive key",
			args: args{
				key: "FOO",
				env: map[string]string{
					"foo": "bar",
				},
			},
			want: "bar",
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			prefix := randSeq(10) + "_"

			for k, v := range tt.args.env {
				key := strings.ToUpper(prefix + k)
				os.Setenv(key, v)
				defer os.Unsetenv(key)
			}

			require.Equal(t, tt.want, env(prefix+tt.args.key))
		})
	}
}

func Test_jsonRequired(t *testing.T) {
	tests := []struct {
		name    string
		val     interface{}
		want    interface{}
		wantErr bool
	}{
		{
			name: "no error",
			val:  true, // any non empty value will do
			want: true,
		},
		{
			name:    "empty item",
			val:     nil,
			wantErr: true,
		},
		{
			name:    "unsupported item",
			val:     struct{ name string }{name: "foo"},
			wantErr: true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got, err := jsonRequired(tt.val)
			requireErrorIf(t, tt.wantErr, err)
			require.Equal(t, tt.want, got)
		})
	}
}

func Test_jsonLowerAndUpper(t *testing.T) {
	type args struct {
		val    interface{}
		prefix string
		suffix string
	}
	tests := []struct {
		name     string
		args     args
		lower    string
		upper    string
		title    string
		trimmed  string
		noprefix string
		nosuffix string
		wantErr  bool
	}{
		{
			name: "generic first test  ",
			args: args{
				val:    "foo bar baz  ",
				prefix: "foo ",
				suffix: " baz  ",
			},
			lower:    "foo bar baz  ",
			upper:    "FOO BAR BAZ  ",
			title:    "Foo Bar Baz  ",
			trimmed:  "foo bar baz",
			noprefix: "bar baz  ",
			nosuffix: "foo bar",
		},
		{
			name: "invalid value type",
			args: args{
				val: struct{}{},
			},
			wantErr: true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			lowered, err := jsonLower(tt.args.val)
			requireErrorIf(t, tt.wantErr, err)

			uppered, err := jsonUpper(tt.args.val)
			requireErrorIf(t, tt.wantErr, err)

			titled, err := jsonTitle(tt.args.val)
			requireErrorIf(t, tt.wantErr, err)

			trimspaced, err := jsonTrimSpace(tt.args.val)
			requireErrorIf(t, tt.wantErr, err)

			prefixed, err := jsonTrimPrefix(tt.args.prefix, tt.args.val)
			requireErrorIf(t, tt.wantErr, err)

			suffixed, err := jsonTrimSuffix(tt.args.suffix, tt.args.val)
			requireErrorIf(t, tt.wantErr, err)

			require.Equal(t, tt.lower, lowered)
			require.Equal(t, tt.upper, uppered)
			require.Equal(t, tt.title, titled)
			require.Equal(t, tt.trimmed, trimspaced)
			require.Equal(t, tt.noprefix, prefixed)
			require.Equal(t, tt.nosuffix, suffixed)
		})
	}
}

func Test_fnDefault(t *testing.T) {
	type args struct {
		defval interface{}
		val    interface{}
	}

	tests := []struct {
		name    string
		args    args
		want    string
		wantErr bool
	}{
		{
			name: "non value use default",
			args: args{
				defval: "foo",
				val:    nil,
			},
			want: "foo",
		},
		{
			name: "existent value skip default",
			args: args{
				defval: "foo",
				val:    "bar",
			},
			want: "bar",
		},
		{
			name: "inconvertible value type use default",
			args: args{
				val: []struct{}{},
			},
			wantErr: true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got, err := fnDefault(tt.args.defval, tt.args.val)
			requireErrorIf(t, tt.wantErr, err)
			require.Equal(t, tt.want, got)
		})
	}
}

func Test_altStrJSON(t *testing.T) {
	tests := []struct {
		name    string
		val     interface{}
		want    string
		wantErr bool
	}{
		{
			name: "default",
			val:  "foo",
			want: "foo\n",
		},
		{
			name: "convert to object",
			val: map[string]interface{}{
				"foo": "bar",
			},
			want: "foo: bar\n",
		},
		{
			name: "convert to array",
			val: []interface{}{
				"foo",
				"bar",
			},
			want: "- foo\n- bar\n",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got, err := altStrJSON(tt.val)
			requireErrorIf(t, tt.wantErr, err)
			require.Equal(t, tt.want, got)
		})
	}
}

func Test_sha256sum(t *testing.T) {
	tests := []struct {
		name    string
		input   interface{}
		want    string
		wantErr bool
	}{
		{
			name:  "generic string",
			input: "foo",
			want:  "b5bb9d8014a0f9b1d61e21e796d78dccdf1352f23cd32812f4850b878ae4944c",
		},
		{
			name: "generic array",
			input: []interface{}{
				"foo",
				"bar",
			},
			want: "d50869a9dcda5fe0b6413eb366dec11d0eb7226c5569f7de8dad1fcd917e5480",
		},
		{
			name: "generic object",
			input: map[string]interface{}{
				"foo": "bar",
			},
			want: "1dabc4e3cbbd6a0818bd460f3a6c9855bfe95d506c74726bc0f2edb0aecb1f4e",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got, err := sha256sum(tt.input)
			requireErrorIf(t, tt.wantErr, err)
			require.Equal(t, tt.want, got)
		})
	}
}

func Test_sha1sum(t *testing.T) {
	tests := []struct {
		name    string
		input   interface{}
		want    string
		wantErr bool
	}{
		{
			name:  "generic string",
			input: "foo",
			want:  "f1d2d2f924e986ac86fdf7b36c94bcdf32beec15",
		},
		{
			name: "generic array",
			input: []interface{}{
				"foo",
				"bar",
			},
			want: "c11e6a294774caece9f882726f0f85c72691bb19",
		},
		{
			name: "generic object",
			input: map[string]interface{}{
				"foo": "bar",
			},
			want: "7e109797e472ae8cbd20d7a4d7e231a96324377c",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got, err := sha1sum(tt.input)
			requireErrorIf(t, tt.wantErr, err)
			require.Equal(t, tt.want, got)
		})
	}
}

func requireErrorIf(t *testing.T, wantErr bool, err error) {
	if wantErr {
		require.Error(t, err)
	} else {
		require.NoError(t, err)
	}
}

func Test_namespaced(t *testing.T) {
	tests := []struct {
		name    string
		input   map[string]interface{}
		want    bool
		wantErr bool
	}{
		{
			name: "builtin cluster scoped",
			input: map[string]interface{}{
				"apiVersion": "v1",
				"kind":       "Namespace",
				"metadata": map[string]interface{}{
					"name": "test",
				},
			},
			want:    false,
			wantErr: false,
		},
		{
			name: "builtin cluster scoped with namespace",
			input: map[string]interface{}{
				"apiVersion": "v1",
				"kind":       "Namespace",
				"metadata": map[string]interface{}{
					"name":      "test",
					"namespace": "test",
				},
			},
			want:    false,
			wantErr: false,
		},
		{
			name: "builtin namespaced",
			input: map[string]interface{}{
				"apiVersion": "apps/v1",
				"kind":       "Deployment",
				"metadata": map[string]interface{}{
					"name":      "test",
					"namespace": "test",
				},
			},
			want:    true,
			wantErr: false,
		},
		{
			name: "builtin namespaced without namespace",
			input: map[string]interface{}{
				"apiVersion": "apps/v1",
				"kind":       "Deployment",
				"metadata": map[string]interface{}{
					"name": "test",
				},
			},
			want:    true,
			wantErr: false,
		},
		{
			name: "generic object with namespace",
			input: map[string]interface{}{
				"apiVersion": "generic/v1",
				"kind":       "Generic",
				"metadata": map[string]interface{}{
					"name":      "test",
					"namespace": "test",
				},
			},
			want:    true,
			wantErr: false,
		},
		{
			name: "generic object without namespace",
			input: map[string]interface{}{
				"apiVersion": "generic/v1",
				"kind":       "Generic",
				"metadata": map[string]interface{}{
					"name": "test",
				},
			},
			want:    false,
			wantErr: false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got, err := namespaced(tt.input)
			requireErrorIf(t, tt.wantErr, err)
			require.Equal(t, tt.want, got)
		})
	}
}
