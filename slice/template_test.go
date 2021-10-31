package slice

import (
	"log"
	"math/rand"
	"os"
	"reflect"
	"strings"
	"testing"
	"time"
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

func Test_mapValueByIndex(t *testing.T) {
	type args struct {
		index string
		m     map[string]interface{}
	}
	tests := []struct {
		name    string
		args    args
		want    interface{}
		wantErr bool
	}{
		{
			name: "fetch existent field",
			args: args{
				index: "foo",
				m: map[string]interface{}{
					"foo": "bar",
				},
			},
			want: "bar",
		},
		{
			name: "fetch nonexistent field",
			args: args{
				index: "baz",
				m: map[string]interface{}{
					"foo": "bar",
				},
			},
			wantErr: true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got, err := mapValueByIndex(tt.args.index, tt.args.m)
			if (err != nil) != tt.wantErr {
				t.Errorf("mapValueByIndex() error = %v, wantErr %v", err, tt.wantErr)
				return
			}
			if !reflect.DeepEqual(got, tt.want) {
				t.Errorf("mapValueByIndex() = %v, want %v", got, tt.want)
			}
		})
	}
}

func Test_strJSON(t *testing.T) {
	type args struct{ val interface{} }

	tests := []struct {
		name    string
		args    args
		want    string
		wantErr bool
	}{
		{
			name: "string conversion",
			args: args{
				val: "foo",
			},
			want: "foo",
		},
		{
			name: "bool true conversion",
			args: args{
				val: true,
			},
			want: "true",
		},
		{
			name: "bool false conversion",
			args: args{
				val: false,
			},
			want: "false",
		},
		{
			name: "float64 conversion",
			args: args{
				val: 3.141592654,
			},
			want: "3.141592654",
		},
		{
			name: "incorrect data type conversion",
			args: args{
				val: []string{},
			},
			wantErr: true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got, err := strJSON(tt.args.val)
			if (err != nil) != tt.wantErr {
				t.Errorf("strJSON() error = %v, wantErr %v", err, tt.wantErr)
				return
			}
			if got != tt.want {
				t.Errorf("strJSON() = %v, want %v", got, tt.want)
			}
		})
	}
}

func Test_jsonAlphanumify(t *testing.T) {
	type args struct{ val interface{} }

	tests := []struct {
		name    string
		args    args
		want    string
		wantErr bool
	}{
		{
			name: "remove dots",
			args: args{
				val: "foo.bar",
			},
			want: "foobar",
		},
		{
			name: "remove dots and slashes",
			args: args{
				val: "foo.bar/baz",
			},
			want: "foobarbaz",
		},
		{
			name: "remove all special characters",
			args: args{
				val: "foo.bar/baz!@#$%^&*()_+-=[]{}\\|;:'\",<.>/?",
			},
			want: "foobarbaz",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got, err := jsonAlphanumify(tt.args.val)
			if (err != nil) != tt.wantErr {
				t.Errorf("jsonAlphanumify() error = %v, wantErr %v", err, tt.wantErr)
				return
			}
			if got != tt.want {
				t.Errorf("jsonAlphanumify() = %v, want %v", got, tt.want)
			}
		})
	}
}

func Test_jsonAlphanumdash(t *testing.T) {
	type args struct{ val interface{} }

	tests := []struct {
		name    string
		args    args
		want    string
		wantErr bool
	}{
		{
			name: "remove dots",
			args: args{
				val: "foo.bar-baz",
			},
			want: "foobar-baz",
		},
		{
			name: "remove dots and slashes",
			args: args{
				val: "foo.bar/baz-daz",
			},
			want: "foobarbaz-daz",
		},
		{
			name: "remove all special characters",
			args: args{
				val: "foo.bar/baz!@#$%^&*()_+=[]{}\\|;:'\",<.>/?-daz",
			},
			want: "foobarbaz-daz",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got, err := jsonAlphanumdash(tt.args.val)
			if (err != nil) != tt.wantErr {
				t.Errorf("jsonAlphanumdash() error = %v, wantErr %v", err, tt.wantErr)
				return
			}
			if got != tt.want {
				t.Errorf("jsonAlphanumdash() = %v, want %v", got, tt.want)
			}
		})
	}
}

func Test_jsonDotToDash(t *testing.T) {
	type args struct{ val interface{} }

	tests := []struct {
		name    string
		args    args
		want    string
		wantErr bool
	}{
		{
			name: "single dot",
			args: args{
				val: "foo.bar",
			},
			want: "foo-bar",
		},
		{
			name: "multi dot",
			args: args{
				val: "foo...bar",
			},
			want: "foo---bar",
		},
		{
			name: "no dot",
			args: args{
				val: "foobar",
			},
			want: "foobar",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got, err := jsonDotToDash(tt.args.val)
			if (err != nil) != tt.wantErr {
				t.Errorf("jsonDotToDash() error = %v, wantErr %v", err, tt.wantErr)
				return
			}
			if got != tt.want {
				t.Errorf("jsonDotToDash() = %v, want %v", got, tt.want)
			}
		})
	}
}

func Test_jsonDotToUnder(t *testing.T) {
	type args struct{ val interface{} }

	tests := []struct {
		name    string
		args    args
		want    string
		wantErr bool
	}{
		{
			name: "single dot",
			args: args{
				val: "foo.bar",
			},
			want: "foo_bar",
		},
		{
			name: "multi dot",
			args: args{
				val: "foo...bar",
			},
			want: "foo___bar",
		},
		{
			name: "no dot",
			args: args{
				val: "foobar",
			},
			want: "foobar",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got, err := jsonDotToUnder(tt.args.val)
			if (err != nil) != tt.wantErr {
				t.Errorf("jsonDotToUnder() error = %v, wantErr %v", err, tt.wantErr)
				return
			}
			if got != tt.want {
				t.Errorf("jsonDotToUnder() = %v, want %v", got, tt.want)
			}
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
			if (err != nil) != tt.wantErr {
				t.Errorf("jsonReplace() error = %v, wantErr %v", err, tt.wantErr)
				return
			}
			if got != tt.want {
				t.Errorf("jsonReplace() = %v, want %v", got, tt.want)
			}
		})
	}
}

func Test_env(t *testing.T) {
	var letters = []rune("abcdefghijklmnopqrstuvwxyz")

	randSeq := func(n int) string {
		b := make([]rune, n)
		for i := range b {
			b[i] = letters[rand.Intn(len(letters))]
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
			rand.Seed(time.Now().UnixNano())
			prefix := randSeq(10) + "_"

			for k, v := range tt.args.env {
				key := strings.ToUpper(prefix + k)
				os.Setenv(key, v)
				defer os.Unsetenv(key)
			}

			if got := env(prefix + tt.args.key); got != tt.want {
				t.Errorf("env() = %v, want %v", got, tt.want)
			}
		})
	}
}

func Test_jsonRequired(t *testing.T) {
	type args struct{ val interface{} }

	tests := []struct {
		name    string
		args    args
		want    interface{}
		wantErr bool
	}{
		{
			name: "no error",
			args: args{
				val: true, // any non empty value will do
			},
			want: true,
		},
		{
			name: "empty item",
			args: args{
				val: nil,
			},
			wantErr: true,
		},
		{
			name: "unsupported item",
			args: args{
				val: struct{ name string }{name: "foo"},
			},
			wantErr: true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got, err := jsonRequired(tt.args.val)

			if (err != nil) != tt.wantErr {
				t.Errorf("jsonRequired() error = %v, wantErr %v", err, tt.wantErr)
				return
			}

			if !reflect.DeepEqual(got, tt.want) {
				t.Errorf("jsonRequired() = %v, want %v", got, tt.want)
			}
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
			if (err != nil) != tt.wantErr {
				t.Errorf("jsonLower() error = %v, wantErr %v", err, tt.wantErr)
				return
			}

			uppered, err := jsonUpper(tt.args.val)
			if (err != nil) != tt.wantErr {
				t.Errorf("jsonUpper() error = %v, wantErr %v", err, tt.wantErr)
				return
			}

			titled, err := jsonTitle(tt.args.val)
			if (err != nil) != tt.wantErr {
				t.Errorf("jsonTitle() error = %v, wantErr %v", err, tt.wantErr)
				return
			}

			trimspaced, err := jsonTrimSpace(tt.args.val)
			if (err != nil) != tt.wantErr {
				t.Errorf("jsonTrimSpace() error = %v, wantErr %v", err, tt.wantErr)
				return
			}

			prefixed, err := jsonTrimPrefix(tt.args.prefix, tt.args.val)
			if (err != nil) != tt.wantErr {
				t.Errorf("jsonTrimPrefix() error = %v, wantErr %v", err, tt.wantErr)
				return
			}

			suffixed, err := jsonTrimSuffix(tt.args.suffix, tt.args.val)
			if (err != nil) != tt.wantErr {
				t.Errorf("jsonTrimSuffix() error = %v, wantErr %v", err, tt.wantErr)
				return
			}

			if lowered != tt.lower {
				t.Errorf("jsonLower() = %v, lower %v", lowered, tt.lower)
			}

			if uppered != tt.upper {
				t.Errorf("jsonUpper() = %v, upper %v", uppered, tt.upper)
			}

			if titled != tt.title {
				t.Errorf("jsonTitle() = %v, title %v", titled, tt.title)
			}

			if trimspaced != tt.trimmed {
				t.Errorf("jsonTrimSpace() = %v, trimspace %v", trimspaced, tt.trimmed)
			}

			if prefixed != tt.noprefix {
				t.Errorf("jsonTrimPrefix() = %v, trimprefix %v", prefixed, tt.noprefix)
			}

			if suffixed != tt.nosuffix {
				t.Errorf("jsonTrimSuffix() = %v, trimsuffix %v", suffixed, tt.nosuffix)
			}
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

			if (err != nil) != tt.wantErr {
				t.Errorf("fnDefault() error = %v, wantErr %v", err, tt.wantErr)
				return
			}

			if got != tt.want {
				t.Errorf("fnDefault() = %v, want %v", got, tt.want)
			}
		})
	}
}

func Test_altStrJSON(t *testing.T) {
	type args struct{ val interface{} }

	tests := []struct {
		name    string
		args    args
		want    string
		wantErr bool
	}{
		{
			name: "default",
			args: args{
				val: "foo",
			},
			want: "foo\n",
		},
		{
			name: "convert to object",
			args: args{
				val: map[string]interface{}{
					"foo": "bar",
				},
			},
			want: "foo: bar\n",
		},
		{
			name: "convert to array",
			args: args{
				val: []interface{}{
					"foo",
					"bar",
				},
			},
			want: "- foo\n- bar\n",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got, err := altStrJSON(tt.args.val)

			if (err != nil) != tt.wantErr {
				t.Errorf("altStrJSON() error = %v, wantErr %v", err, tt.wantErr)
				return
			}

			if got != tt.want {
				t.Errorf("altStrJSON() = %q, want %q", got, tt.want)
			}
		})
	}
}

func Test_sha256sum(t *testing.T) {
	type args struct{ input interface{} }

	tests := []struct {
		name    string
		args    args
		want    string
		wantErr bool
	}{
		{
			name: "generic string",
			args: args{
				input: "foo",
			},
			want: "b5bb9d8014a0f9b1d61e21e796d78dccdf1352f23cd32812f4850b878ae4944c",
		},
		{
			name: "generic array",
			args: args{
				input: []interface{}{
					"foo",
					"bar",
				},
			},
			want: "d50869a9dcda5fe0b6413eb366dec11d0eb7226c5569f7de8dad1fcd917e5480",
		},
		{
			name: "generic object",
			args: args{
				input: map[string]interface{}{
					"foo": "bar",
				},
			},
			want: "1dabc4e3cbbd6a0818bd460f3a6c9855bfe95d506c74726bc0f2edb0aecb1f4e",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got, err := sha256sum(tt.args.input)

			if (err != nil) != tt.wantErr {
				t.Errorf("sha256sum() error = %v, wantErr %v", err, tt.wantErr)
				return
			}

			if got != tt.want {
				t.Errorf("sha256sum() = %v, want %v", got, tt.want)
			}
		})
	}
}

func Test_sha1sum(t *testing.T) {
	type args struct{ input interface{} }

	tests := []struct {
		name    string
		args    args
		want    string
		wantErr bool
	}{
		{
			name: "generic string",
			args: args{
				input: "foo",
			},
			want: "f1d2d2f924e986ac86fdf7b36c94bcdf32beec15",
		},
		{
			name: "generic array",
			args: args{
				input: []interface{}{
					"foo",
					"bar",
				},
			},
			want: "c11e6a294774caece9f882726f0f85c72691bb19",
		},
		{
			name: "generic object",
			args: args{
				input: map[string]interface{}{
					"foo": "bar",
				},
			},
			want: "7e109797e472ae8cbd20d7a4d7e231a96324377c",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got, err := sha1sum(tt.args.input)

			if (err != nil) != tt.wantErr {
				t.Errorf("sha1sum() error = %v, wantErr %v", err, tt.wantErr)
				return
			}

			if got != tt.want {
				t.Errorf("sha1sum() = %v, want %v", got, tt.want)
			}
		})
	}
}
