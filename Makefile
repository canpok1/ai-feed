BINARY_NAME=ai-feed
VERSION?=dev
COVERAGE_THRESHOLD=60

setup:
	go install go.uber.org/mock/mockgen@v0.6.0
	go install golang.org/x/tools/cmd/goimports@v0.28.0
	go install github.com/golangci/golangci-lint/cmd/golangci-lint@v1.64.8

run:
	@go run main.go ${option}

build:
	go build -ldflags "-X github.com/canpok1/ai-feed/cmd.version=${VERSION}" -o ${BINARY_NAME} main.go

build-release:
	goreleaser build --snapshot --clean

clean:
	go clean
	rm -f ${BINARY_NAME}
	rm -rf ./dist
	rm -f coverage.out coverage.html

test:
	go test ./...

test-integration:
	go test -tags=integration -v ./cmd/...

test-performance:
	go test -tags=integration -v -run="Performance" ./cmd/...

test-all: test test-integration

test-coverage:
	go test -coverprofile=coverage.out ./...
	go tool cover -html=coverage.out -o coverage.html

test-coverage-check:
	@go test -coverprofile=coverage.out ./...
	@go tool cover -func=coverage.out | grep total | awk '{print $$3}' | sed 's/%//' | \
	awk -v thold=$(COVERAGE_THRESHOLD) '{if ($$1 < thold) {printf "Coverage %.2f%% is below threshold %d%%\n", $$1, thold; exit 1} else {printf "Coverage %.2f%% meets threshold %d%%\n", $$1, thold}}'

lint:
	go vet ./...

lint-all:
	golangci-lint run ./...

fmt:
	go fmt ./...
	export PATH=$$PATH:$$(go env GOPATH)/bin && go list -f '{{.Dir}}' ./... | xargs goimports -w

generate:
	go generate ./...
