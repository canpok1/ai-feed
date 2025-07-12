BINARY_NAME=ai-feed

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
	go test -v ./...
