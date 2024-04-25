## Go builder image
FROM golang:1.20.3 as builder

ENV GOPATH=/usr/local/go/bin

WORKDIR /go

COPY . .

RUN CGO_ENABLED=0 go build -ldflags "-s" -a -installsuffix cgo .

## Base image for production usage
FROM alpine:3.14 as production

COPY --from=builder /go/kubectl-slice /usr/bin/kubectl-slice

WORKDIR /workdir

## add user k8s to run the slice tool
RUN set -eux; \
  addgroup -g 1000 k8s; \
  adduser -u 1000 -G k8s -s /bin/sh -h /home/k8s -D k8s

RUN chown -R k8s:k8s /workdir

USER k8s

ENTRYPOINT ["-c", "/usr/bin/kubectl-slice"]