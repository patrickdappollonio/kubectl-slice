package template

import (
	"math/rand/v2"
	"os"
	"testing"

	"github.com/stretchr/testify/require"
)

func Test_mapValueByIndexEmpty(t *testing.T) {
	tests := []struct {
		name  string
		index string
		m     map[string]any
		want  any
	}{
		{
			name:  "nil map",
			index: "foo",
			m:     nil,
			want:  "",
		},
		{
			name:  "empty index",
			index: "",
			m:     map[string]any{},
			want:  "",
		},
		{
			name:  "key not found",
			index: "foo",
			m:     map[string]any{"bar": "baz"},
			want:  "",
		},
		{
			name:  "key found",
			index: "foo",
			m:     map[string]any{"foo": "bar"},
			want:  "bar",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got := mapValueByIndexOrEmpty(tt.index, tt.m)
			require.Equal(t, tt.want, got)
		})
	}
}

func Test_mapValueByIndex(t *testing.T) {
	tests := []struct {
		name    string
		index   string
		m       map[string]any
		want    any
		wantErr bool
	}{
		{
			name:    "nil map",
			index:   "foo",
			m:       nil,
			want:    nil,
			wantErr: true,
		},
		{
			name:    "empty index",
			index:   "",
			m:       map[string]any{},
			want:    nil,
			wantErr: true,
		},
		{
			name:    "key not found",
			index:   "foo",
			m:       map[string]any{"bar": "baz"},
			want:    nil,
			wantErr: true,
		},
		{
			name:    "key found",
			index:   "foo",
			m:       map[string]any{"foo": "bar"},
			want:    "bar",
			wantErr: false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got, err := mapValueByIndex(tt.index, tt.m)
			if tt.wantErr {
				require.Error(t, err)
				return
			}

			require.NoError(t, err)
			require.Equal(t, tt.want, got)
		})
	}
}

func Test_jsonLower(t *testing.T) {
	require.Equal(t, "foo", jsonLower("FOO"))
	require.Equal(t, "foo", jsonLower("foo"))
	require.Equal(t, "123", jsonLower(123))
	require.Equal(t, "", jsonLower(nil))
}

func Test_jsonUpper(t *testing.T) {
	require.Equal(t, "FOO", jsonUpper("foo"))
	require.Equal(t, "FOO", jsonUpper("FOO"))
	require.Equal(t, "123", jsonUpper(123))
	require.Equal(t, "", jsonUpper(nil))
}

func Test_jsonTitle(t *testing.T) {
	require.Equal(t, "Foo", jsonTitle("foo"))
	require.Equal(t, "Foo", jsonTitle("FOO"))
	require.Equal(t, "123", jsonTitle(123))
	require.Equal(t, "", jsonTitle(nil))
}

func Test_jsonTrimSpace(t *testing.T) {
	require.Equal(t, "foo", jsonTrimSpace(" foo "))
	require.Equal(t, "foo", jsonTrimSpace("foo"))
	require.Equal(t, "123", jsonTrimSpace(123))
	require.Equal(t, "", jsonTrimSpace(nil))
}

func Test_jsonTrimPrefix(t *testing.T) {
	require.Equal(t, "bar", jsonTrimPrefix("foo", "foobar"))
	require.Equal(t, "bar", jsonTrimPrefix("foo", "bar"))
	require.Equal(t, "123", jsonTrimPrefix("foo", 123))
	require.Equal(t, "", jsonTrimPrefix("foo", nil))
}

func Test_jsonTrimSuffix(t *testing.T) {
	require.Equal(t, "foo", jsonTrimSuffix("bar", "foobar"))
	require.Equal(t, "foo", jsonTrimSuffix("bar", "foo"))
	require.Equal(t, "123", jsonTrimSuffix("bar", 123))
	require.Equal(t, "", jsonTrimSuffix("bar", nil))
}

func Test_fnDefault(t *testing.T) {
	require.Equal(t, "foo", fnDefault("foo", ""))
	require.Equal(t, "bar", fnDefault("foo", "bar"))
	require.Equal(t, "123", fnDefault("foo", 123))
	require.Equal(t, "foo", fnDefault("foo", nil))
}

func Test_sha1sum(t *testing.T) {
	require.Equal(t, "0beec7b5ea3f0fdbc95d0dd47f3c5bc275da8a33", sha1sum("foo"))
	require.Equal(t, "da39a3ee5e6b4b0d3255bfef95601890afd80709", sha1sum(""))
	require.Equal(t, "da39a3ee5e6b4b0d3255bfef95601890afd80709", sha1sum(nil))
}

func Test_sha256sum(t *testing.T) {
	require.Equal(t, "2c26b46b68ffc68ff99b453c1d30413413422d706483bfa0f98a5e886266e7ae", sha256sum("foo"))
	require.Equal(t, "e3b0c44298fc1c149afbf4c8996fb92427ae41e4649b934ca495991b7852b855", sha256sum(""))
	require.Equal(t, "e3b0c44298fc1c149afbf4c8996fb92427ae41e4649b934ca495991b7852b855", sha256sum(nil))
}

func Test_strJSON(t *testing.T) {
	require.Equal(t, "foo\n", strJSON("foo"))
	require.Equal(t, "123\n", strJSON(123))
	require.Equal(t, "null\n", strJSON(nil))
}

func Test_jsonRequired(t *testing.T) {
	v, err := jsonRequired("value is required", "foo")
	require.NoError(t, err)
	require.Equal(t, "foo", v)

	v, err = jsonRequired("value is required", "")
	require.Error(t, err)
	require.Equal(t, "", v)

	v, err = jsonRequired("value is required", nil)
	require.Error(t, err)
	require.Equal(t, nil, v)
}

func Test_env(t *testing.T) {
	os.Setenv("TEST_ENV_VAR", "foo")
	defer os.Unsetenv("TEST_ENV_VAR")

	require.Equal(t, "foo", env("TEST_ENV_VAR"))
	require.Equal(t, "", env("NONEXISTENT_ENV_VAR"))
	require.Equal(t, "", env(""))
}

func Test_jsonReplace(t *testing.T) {
	require.Equal(t, "foobaz", jsonReplace("bar", "baz", "foobar"))
	require.Equal(t, "123", jsonReplace("bar", "baz", 123))
	require.Equal(t, "", jsonReplace("bar", "baz", nil))
}

func Test_jsonAlphanumify(t *testing.T) {
	require.Equal(t, "foobar123", jsonAlphanumify("foo-bar-123"))
	require.Equal(t, "foobar123", jsonAlphanumify("foo_bar_123"))
	require.Equal(t, "foobar123", jsonAlphanumify("foo.bar.123"))
	require.Equal(t, "123", jsonAlphanumify(123))
	require.Equal(t, "", jsonAlphanumify(nil))
}

func Test_jsonAlphanumdash(t *testing.T) {
	require.Equal(t, "foo-bar-123", jsonAlphanumdash("foo-bar-123"))
	require.Equal(t, "foo-bar-123", jsonAlphanumdash("foo_bar_123"))
	require.Equal(t, "foo-bar-123", jsonAlphanumdash("foo.bar.123"))
	require.Equal(t, "123", jsonAlphanumdash(123))
	require.Equal(t, "", jsonAlphanumdash(nil))
}

func Test_jsonDotToDash(t *testing.T) {
	require.Equal(t, "foo-bar-123", jsonDotToDash("foo.bar.123"))
	require.Equal(t, "foo_bar_123", jsonDotToDash("foo_bar_123"))
	require.Equal(t, "123", jsonDotToDash(123))
	require.Equal(t, "", jsonDotToDash(nil))
}

func Test_jsonDotToUnder(t *testing.T) {
	require.Equal(t, "foo_bar_123", jsonDotToUnder("foo.bar.123"))
	require.Equal(t, "foo-bar-123", jsonDotToUnder("foo-bar-123"))
	require.Equal(t, "123", jsonDotToUnder(123))
	require.Equal(t, "", jsonDotToUnder(nil))
}

func Test_Pluralize(t *testing.T) {
	require.Equal(t, "foo", Pluralize("foo", 1))
	require.Equal(t, "foos", Pluralize("foo", 0))
	require.Equal(t, "foos", Pluralize("foo", 2))
}

func Test_toString(t *testing.T) {
	require.Equal(t, "foo", toString("foo"))
	require.Equal(t, "foo", toString([]byte("foo")))
	require.Equal(t, "123", toString(123))
	require.Equal(t, "", toString(nil))

	// Test error type
	errMsg := randomString(10)
	require.Equal(t, errMsg, toString(errorString(errMsg)))

	// Test fmt.Stringer
	stringerMsg := randomString(10)
	require.Equal(t, stringerMsg, toString(stringStringer(stringerMsg)))
}

type errorString string

func (e errorString) Error() string {
	return string(e)
}

type stringStringer string

func (s stringStringer) String() string {
	return string(s)
}

func randomString(n int) string {
	const letterBytes = "abcdefghijklmnopqrstuvwxyzABCDEFGHIJKLMNOPQRSTUVWXYZ"
	b := make([]byte, n)
	for i := range b {
		b[i] = letterBytes[int(rand.Int32N(int32(len(letterBytes))))]
	}
	return string(b)
}
