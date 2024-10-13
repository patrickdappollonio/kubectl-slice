ARG KUBECTL_VERSION=1.31.1
ARG YQ_VERSION=v4.44.3

# Stage 1: Download binaries
FROM alpine as download_binary

ARG KUBECTL_VERSION
ARG YQ_VERSION

# Install curl and certificates, and clean up in one layer to reduce image size
RUN apk update && apk add --no-cache \
    curl \
    ca-certificates

# Download kubectl binary
RUN curl -sSL -o /kubectl "https://dl.k8s.io/release/v${KUBECTL_VERSION}/bin/linux/amd64/kubectl" \
  && chmod +x /kubectl

# Download yq binary
RUN curl -sSL -o /yq "https://github.com/mikefarah/yq/releases/download/${YQ_VERSION}/yq_linux_amd64" \
  && chmod +x /yq

# Stage 2
FROM alpine 

RUN apk update && apk add --no-cache \
  sudo \
  && adduser -D -s /bin/bash slice \
  && echo 'slice ALL=(ALL) NOPASSWD:ALL' > /etc/sudoers.d/slice \
  && chmod 0440 /etc/sudoers.d/slice 

# Copy binaries from the download_binary stage
COPY --from=download_binary /kubectl /usr/local/bin/kubectl
COPY --from=download_binary /yq /usr/local/bin/yq
COPY --from=download_binary /etc/ssl/certs/ca-certificates.crt /etc/ssl/certs/ca-certificates.crt

# Copy kubectl-slice from local filesystem
COPY kubectl-slice /usr/local/bin/kubectl-slice
USER slice
WORKDIR /home/slice
