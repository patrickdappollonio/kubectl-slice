package slice

import (
	"fmt"
	"io/ioutil"
	"os"
	"strings"
	"testing"
)

const fileFolderPrefix = "kubectl-slice-test-"

var largeFile = `
---
# foo
---
---
# apiVersion: v1
# kind: Pod
# metadata:
#   name: hello-docker
#   labels:
#     app: hello-docker-app
# spec:
#   containers:
#   - name: hello-docker-container
#     image: patrickdappollonio/hello-docker

---

apiVersion: apps/v1
kind: Deployment
metadata:
  name: hello-docker
spec:
  replicas: 3
  selector:
    matchLabels:
      app: hello-docker-app
  template:
    metadata:
      labels:
        app: hello-docker-app
    spec:
      containers:
      - name: hello-docker-container
        image: patrickdappollonio/hello-docker

---

apiVersion: v1
kind: Service
metadata:
  name: hello-docker-svc
spec:
  selector:
    app: hello-docker-app
  ports:
  - port: 8000

---

apiVersion: networking.k8s.io/v1
kind: Ingress
metadata:
  name: hello-docker-ing
spec:
  ingressClassName: nginx
  rules:
  - host: foo.bar
    http:
      paths:
      - path: /
        pathType: Prefix
        backend:
          service:
            name: hello-docker-svc
            port:
              number: 8000`

var largeFileSplitted = map[string]string{
	"deployment-hello-docker.yaml": `apiVersion: apps/v1
kind: Deployment
metadata:
  name: hello-docker
spec:
  replicas: 3
  selector:
    matchLabels:
      app: hello-docker-app
  template:
    metadata:
      labels:
        app: hello-docker-app
    spec:
      containers:
      - name: hello-docker-container
        image: patrickdappollonio/hello-docker`,
	"ingress-hello-docker-ing.yaml": `apiVersion: networking.k8s.io/v1
kind: Ingress
metadata:
  name: hello-docker-ing
spec:
  ingressClassName: nginx
  rules:
  - host: foo.bar
    http:
      paths:
      - path: /
        pathType: Prefix
        backend:
          service:
            name: hello-docker-svc
            port:
              number: 8000`,
	"service-hello-docker-svc.yaml": `apiVersion: v1
kind: Service
metadata:
  name: hello-docker-svc
spec:
  selector:
    app: hello-docker-app
  ports:
  - port: 8000`,
	"-.yaml": `# foo

---

# apiVersion: v1
# kind: Pod
# metadata:
#   name: hello-docker
#   labels:
#     app: hello-docker-app
# spec:
#   containers:
#   - name: hello-docker-container
#     image: patrickdappollonio/hello-docker`,
}

var allInOne = `foo: bar
---
foo: baz
---
bar: baz`

var allInOneOutput = map[string]string{
	"example.yaml": `foo: bar

---

foo: baz

---

bar: baz`,
}

func TestEndToEnd(t *testing.T) {
	cases := []struct {
		name          string
		inputFile     string
		template      string
		expectedFiles map[string]string
	}{
		{
			name:          "end to end",
			inputFile:     largeFile,
			template:      DefaultTemplateName,
			expectedFiles: largeFileSplitted,
		},
		{
			name:          "everything in a single file",
			inputFile:     allInOne,
			template:      "example.yaml",
			expectedFiles: allInOneOutput,
		},
	}

	createTempYAML := func(contents string) (string, error) {
		f, err := ioutil.TempFile("/tmp", fileFolderPrefix+"file-*")
		if err != nil {
			return "", fmt.Errorf("unable to create temporary file: %w", err)
		}

		fmt.Fprintln(f, contents)
		f.Close()

		return f.Name(), nil
	}

	for _, v := range cases {
		t.Run(v.name, func(tt *testing.T) {
			temp, err := createTempYAML(v.inputFile)
			if err != nil {
				tt.Fatal(err.Error())
			}

			defer os.Remove(temp)

			dir, err := ioutil.TempDir("/tmp", fileFolderPrefix+"folder-*")
			if err != nil {
				tt.Fatalf("unable to create temporary directory: %s", err.Error())
			}

			defer os.RemoveAll(dir)

			opts := Options{
				InputFile:       temp,
				OutputDirectory: dir,
				GoTemplate:      v.template,
			}
			slice, err := New(opts)

			t.Logf("Opts: %#v", opts)

			if err != nil {
				tt.Fatalf("not expecting an error on New() but got: %s", err.Error())
			}

			if err := slice.Execute(); err != nil {
				tt.Fatalf("not expecting an error on Execute() but got: %s", err.Error())
			}

			files, err := ioutil.ReadDir(dir)
			if err != nil {
				tt.Fatalf("unable to read directory %q: %s", dir, err.Error())
			}

			converted := make(map[string]string)
			for _, v := range files {
				if v.IsDir() {
					continue
				}

				if !strings.HasSuffix(v.Name(), ".yaml") {
					continue
				}

				f, err := ioutil.ReadFile(dir + "/" + v.Name())
				if err != nil {
					tt.Fatalf("unable to read file %q: %s", v.Name(), err.Error())
				}

				converted[v.Name()] = string(f)
			}

			if a, b := len(v.expectedFiles), len(converted); a != b {
				tt.Fatalf("expecting to get %d files but got %d converted", a, b)
			}

			for originalName, originalValue := range v.expectedFiles {
				originalValue = originalValue + "\n\n"
				found := false
				currentFile := ""

				for receivedName, receivedValue := range converted {
					if originalName == receivedName {
						found = true
						currentFile = receivedValue
						break
					}
				}

				if !found {
					tt.Fatalf("expecting to find file %q but it was not found", originalName)
				}

				if originalValue != currentFile {
					tt.Fatalf(
						"on file %q, expecting to get %d bytes of data, and got %d: files are different\n%s\n%s",
						originalName, len(originalValue), len(currentFile), originalValue, currentFile,
					)
				}
			}
		})
	}
}
