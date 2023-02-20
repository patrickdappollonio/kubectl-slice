package main

import (
	"bytes"
	"os"
	"testing"

	"github.com/stretchr/testify/require"
)

func TestMainApp(t *testing.T) {
	// A quick and dirty test run to make sure the
	// application run without errors and no regressions
	// are generated.
	// We also test for skipping non Kubernetes resources
	// in fake files.

	cases := []struct {
		name   string
		file   string
		flags  []string
		stdout string
		stderr string
	}{
		{
			name: "readme sample",
			file: `apiVersion: v1
kind: Pod
metadata:
  name: nginx-ingress
---
apiVersion: v1
kind: Namespace
metadata:
  name: production`,
			stdout: `# File: pod-nginx-ingress.yaml (56 bytes)
apiVersion: v1
kind: Pod
metadata:
  name: nginx-ingress

---
# File: namespace-production.yaml (59 bytes)
apiVersion: v1
kind: Namespace
metadata:
  name: production
`,
			stderr: `2 files parsed to stdout.`,
		},
		{
			name: "non-kubernetes file",
			file: `kind: foo
name: bar
age: baz
---
another: file`,
			stdout: `# File: foo-.yaml (28 bytes)
kind: foo
name: bar
age: baz

---
# File: -.yaml (13 bytes)
another: file`,
			stderr: `2 files parsed to stdout.`,
		},
		{
			name: "non-kubernetes file with skip non k8s enabled",
			file: `kind: foo
name: bar
age: baz
---
another: file`,
			flags:  []string{"--skip-non-k8s"},
			stdout: "",
			stderr: `0 files parsed to stdout.`,
		},
	}

	for _, c := range cases {
		t.Run(c.name, func(tt *testing.T) {
			f, err := os.CreateTemp("/tmp", "kubectl-slice-testing-file-*")
			require.NoErrorf(tt, err, "unable to create temporary file")
			defer os.Remove(f.Name())

			_, err = f.Write([]byte(c.file))
			require.NoErrorf(tt, err, "unable to write to temporary file")

			var stdout, stderr bytes.Buffer

			baseArgs := []string{"--input-file=" + f.Name(), "--stdout"}
			args := append(baseArgs, c.flags...)

			cmd := root()
			cmd.SetOut(&stdout)
			cmd.SetErr(&stderr)
			cmd.SetArgs(args)
			require.NoError(tt, cmd.Execute())
			require.EqualValues(tt, c.stdout, stdout.String())
			require.EqualValues(tt, c.stderr, stderr.String())
		})
	}
}
