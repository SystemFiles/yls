FROM golang:1.19 as builder
ARG VERSION
ARG TARGETOS
ARG TARGETARCH

WORKDIR /workspace
COPY go.mod go.mod
COPY go.sum go.sum
# cache deps before building and copying source so that we don't need to re-download as much
# and so that source changes don't invalidate our downloaded layer
RUN go mod download

# Copy the source code
COPY main.go main.go
COPY cmd/ cmd/
COPY pkg/ pkg/

# Build
# the GOARCH has not a default value to allow the binary be built according to the host where the command
# was called. For example, if we call make docker-build in a local env which has the Apple Silicon M1 SO
# the docker BUILDPLATFORM arg will be linux/arm64 when for Apple x86 it will be linux/amd64. Therefore,
# by leaving it empty we can ensure that the container and binary shipped on it will have the same platform.
RUN CGO_ENABLED=0 GOOS=${TARGETOS:-linux} GOARCH=${TARGETARCH} go build -a -ldflags="-s -w -X sykesdev.ca/yls/cmd.version=${VERSION} -X sykesdev.ca/yls/cmd.commitSha=${VERSION} -X sykesdev.ca/yls/cmd.targetOs=${TARGETOS} -X sykesdev.ca/yls/cmd.targetArch=${TARGETARCH} -X sykesdev.ca/yls/cmd.buildstamp=$(date +%s)" -o yls main.go

# use scratch to limit attack surface-area
FROM scratch
WORKDIR /
COPY --from=builder /workspace/yls .
USER 65532:65532

ENTRYPOINT ["/yls"]