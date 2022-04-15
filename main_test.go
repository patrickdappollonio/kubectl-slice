package main

import (
	"bytes"
	"os"
	"strings"
	"testing"
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
			stdout: `# File: pod-nginx-ingress.yaml (58 bytes)
apiVersion: v1
kind: Pod
metadata:
  name: nginx-ingress


---

# File: namespace-production.yaml (61 bytes)
apiVersion: v1
kind: Namespace
metadata:
  name: production`,
			stderr: `2 files parsed to stdout.`,
		},
		{
			name: "non-kubernetes file",
			file: `kind: foo
name: bar
age: baz
---
another: file`,
			stdout: `# File: foo-.yaml (30 bytes)
kind: foo
name: bar
age: baz


---

# File: -.yaml (15 bytes)
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
			if err != nil {
				tt.Fatalf("unable to create temporary file: %s", err.Error())
			}
			defer os.Remove(f.Name())

			if _, err := f.Write([]byte(c.file)); err != nil {
				tt.Fatalf("unable to write to temporary file: %s", err.Error())
			}

			var stdout bytes.Buffer
			var stderr bytes.Buffer

			baseArgs := []string{"--input-file=" + f.Name(), "--stdout"}
			args := append(baseArgs, c.flags...)

			cmd := root()
			cmd.SetOut(&stdout)
			cmd.SetErr(&stderr)
			cmd.SetArgs(args)
			if err := cmd.Execute(); err != nil {
				tt.Fatalf("unable to execute command: %s", err.Error())
			}

			if strings.TrimSpace(stdout.String()) != c.stdout {
				tt.Fatalf("stdout mismatch: expected %s, got %s", c.stdout, stdout.String())
			}

			if strings.TrimSpace(stderr.String()) != c.stderr {
				tt.Fatalf("stderr mismatch: expected %s, got %s", c.stderr, stderr.String())
			}
		})
	}
}
