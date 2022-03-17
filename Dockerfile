# Build the manager binary
FROM golang:1.17.7 as builder

WORKDIR /workspace
# Copy the Go Modules manifests
COPY go.mod go.mod
COPY go.sum go.sum
# cache deps before building and copying source so that we don't need to re-download as much
# and so that source changes don't invalidate our downloaded layer
RUN export GOPROXY=https://goproxy.cn  \
&&  go mod download

# Copy the go source
COPY main.go main.go
COPY api/ api/
COPY controllers/ controllers/
COPY iaw-shared-helpers/ iaw-shared-helpers/
COPY internal/ internal/

# Build
RUN CGO_ENABLED=0 GOOS=linux GOARCH=amd64 go build -a -o manager main.go

# Use distroless as minimal base image to package the manager binary
FROM docker.io/redhat/ubi8-minimal:latest
COPY --from=builder /workspace/manager /manager

RUN microdnf install -y shadow-utils \
    && adduser manager  -u 10001 -g 0 \
    && chown manager:root /manager \
    && chmod +x /manager

USER 10001

ENTRYPOINT ["/manager"]




