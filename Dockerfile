# Build the manager binary
FROM golang:1.13 as builder

WORKDIR /workspace

# Copy the go source
COPY . .

# Build
RUN CGO_ENABLED=0 GOOS=linux GOARCH=amd64 GO111MODULE=on make build

# Use distroless as minimal base image to package the manager binary
# Refer to https://github.com/GoogleContainerTools/distroless for more details
FROM gcr.io/distroless/static:nonroot as controller
WORKDIR /
COPY --from=builder /go/bin/secret-replication-controller /secret-replication-controller

USER nonroot:nonroot

ENTRYPOINT ["/secret-replication-controller"]
