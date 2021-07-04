

revendor:
	@go mod vendor
	@go mod tidy

generate:
	@go generate ./...

# Run tests
test:
	@go test ./... -coverprofile cover.out

build: generate fmt vet
	@go install -mod=vendor ./cmd/secret-replication-controller

# Run go fmt against code
fmt:
	@go fmt ./...

# Run go vet against code
vet:
	@go vet ./...
