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
			name:   "readme sample",
			file:   "slice/testdata/ingress-namespace.yaml",
			stdout: "slice/testdata/ingress-namespace/stdout.yaml",
			stderr: "slice/testdata/ingress-namespace/stderr",
		},
		{
			name:   "non-kubernetes file",
			file:   "slice/testdata/non-kubernetes.yaml",
			stdout: "slice/testdata/non-kubernetes/stdout.yaml",
			stderr: "slice/testdata/non-kubernetes/stderr",
		},
		{
			name:   "non-kubernetes file with skip non k8s enabled",
			file:   "slice/testdata/non-kubernetes-skip.yaml",
			flags:  []string{"--skip-non-k8s"},
			stdout: "slice/testdata/non-kubernetes-skip/stdout",
			stderr: "slice/testdata/non-kubernetes-skip/stderr",
		},
	}

	for _, c := range cases {
		t.Run(c.name, func(tt *testing.T) {
			var stdout, stderr bytes.Buffer

			baseArgs := []string{"--input-file=" + c.file, "--stdout"}
			args := append(baseArgs, c.flags...)

			cmd := root()
			cmd.SetOut(&stdout)
			cmd.SetErr(&stderr)
			cmd.SetArgs(args)
			require.NoError(tt, cmd.Execute())

			appout, err := os.ReadFile(c.stdout)
			require.NoError(tt, err)
			apperr, err := os.ReadFile(c.stderr)
			require.NoError(tt, err)

			require.EqualValues(tt, string(appout), stdout.String())
			require.EqualValues(tt, string(apperr), stderr.String())
		})
	}
}
