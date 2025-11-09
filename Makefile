BINARY_NAME=ai-feed
VERSION?=dev

setup:
	go install go.uber.org/mock/mockgen@latest
	go install golang.org/x/tools/cmd/goimports@latest
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
	awk '{if ($$1 < 60) {print "Coverage " $$1 "% is below threshold 60%"; exit 1} else {print "Coverage " $$1 "% meets threshold 60%"}}'

lint:
	go vet ./...

lint-all:
	golangci-lint run ./...

fmt:
	go fmt ./...
	go list -f '{{.Dir}}' ./... | xargs goimports -w

generate:
	go generate ./...
