BINARY_NAME=ai-feed
VERSION?=dev
COVERAGE_THRESHOLD=60

setup:
	go install go.uber.org/mock/mockgen@v0.6.0
	go install golang.org/x/tools/cmd/goimports@v0.28.0
	go install github.com/golangci/golangci-lint/cmd/golangci-lint@v1.64.8
	go install github.com/goreleaser/goreleaser/v2@v2.5.1

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

test-e2e:
	@if [ -z "$$GEMINI_API_KEY" ]; then \
		echo "エラー: GEMINI_API_KEY環境変数が設定されていません"; \
		echo "e2eテストを実行するには、Gemini APIキーが必要です"; \
		echo "設定方法: export GEMINI_API_KEY=your_api_key"; \
		exit 1; \
	fi
	go test -tags=e2e -v ./test/e2e/...

test-coverage:
	@go test -coverprofile=coverage.out ./...
	@go tool cover -html=coverage.out -o coverage.html
	@go tool cover -func=coverage.out | awk -v thold=$(COVERAGE_THRESHOLD) '/^total:/ {gsub(/%/, "", $$3); if ($$3 < thold) {printf "Coverage %.2f%% is below threshold %d%%\n", $$3, thold; exit 1} else {printf "Coverage %.2f%% meets threshold %d%%\n", $$3, thold}}'

lint:
	golangci-lint run ./...

fmt:
	go fmt ./...
	go list -f '{{.Dir}}' ./... | xargs goimports -w

generate:
	go generate ./...
