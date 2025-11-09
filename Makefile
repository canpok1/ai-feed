BINARY_NAME=ai-feed
VERSION?=dev

setup:
	go install go.uber.org/mock/mockgen@latest
	go install golang.org/x/tools/cmd/goimports@latest
	go install github.com/golangci/golangci-lint/cmd/golangci-lint@latest

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

test:
	go test ./...

test-integration:
	go test -tags=integration -v ./cmd/...

test-performance:
	go test -tags=integration -v -run="Performance" ./cmd/...

test-all: test test-integration

lint:
	go vet ./...

lint-all:
	golangci-lint run ./...

fmt:
	go fmt ./...
	go list -f '{{.Dir}}' ./... | xargs goimports -w

generate:
	go generate ./...
