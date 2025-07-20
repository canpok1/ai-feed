BINARY_NAME=ai-feed

setup:
	go install go.uber.org/mock/mockgen@latest

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

lint:
	go vet ./...

fmt:
	go fmt ./...

generate:
	go generate ./...
