BINARY_NAME=ai-feed

setup:
	go install go.uber.org/mock/mockgen@latest
	go install golang.org/x/tools/cmd/goimports@latest

run:
	@go run main.go ${option}

build:
	go build -o ${BINARY_NAME} main.go

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

fmt:
	go fmt ./...
	go list -f '{{.Dir}}' ./... | xargs goimports -w

generate:
	go generate ./...
