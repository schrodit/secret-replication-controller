
install-requirements:
	@go install sigs.k8s.io/controller-runtime/tools/setup-envtest@latest

revendor:
	@go mod vendor
	@go mod tidy

generate:
	@go generate ./...

# Run tests
test:
	@./hack/test.sh

build: generate fmt vet
	@go install -mod=vendor ./cmd/secret-replication-controller

# Run go fmt against code
fmt:
	@go fmt ./...

# Run go vet against code
vet:
	@go vet ./...
