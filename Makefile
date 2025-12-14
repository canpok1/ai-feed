BINARY_NAME=ai-feed
VERSION?=dev
COVERAGE_THRESHOLD_DOMAIN=80
COVERAGE_THRESHOLD_INFRA=60

setup:
	go install go.uber.org/mock/mockgen@v0.6.0
	go install golang.org/x/tools/cmd/goimports@v0.28.0
	go install github.com/golangci/golangci-lint/v2/cmd/golangci-lint@v2.5.0
	go install github.com/goreleaser/goreleaser/v2@v2.5.1
	go install github.com/v-standard/go-depcheck/cmd/depcheck@v0.0.2

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
	rm -f coverage.out coverage.filtered.out coverage.func.out coverage.html
	rm -rf public/coverage

test:
	go test ./...

test-integration:
	go test -tags=integration ./test/integration/...

test-e2e:
	go test -tags=e2e -v ./test/e2e/...

test-coverage:
	@go test -tags=integration -coverprofile=coverage.out -coverpkg=./internal/... ./... ./test/integration/...
	@grep -v "mock_" coverage.out > coverage.filtered.out
	@mkdir -p public/coverage/ut-it
	@go tool cover -html=coverage.filtered.out -o public/coverage/ut-it/index.html
	@go tool cover -func=coverage.filtered.out > coverage.func.out
	@echo "=== Layer Coverage Check (per docs/03_testing_rules.md) ==="
	@# domain層のカバレッジチェック（80%以上）
	@domain_cov=$$(awk '/internal\/domain/ {gsub(/%/, "", $$NF); sum+=$$NF; count++} END {if(count>0) printf "%.1f", sum/count; else print "0"}' coverage.func.out); \
	echo "domain layer: $${domain_cov}% (threshold: $(COVERAGE_THRESHOLD_DOMAIN)%)"; \
	if [ $$(echo "$${domain_cov} < $(COVERAGE_THRESHOLD_DOMAIN)" | bc -l) -eq 1 ]; then \
		echo "ERROR: domain layer coverage $${domain_cov}% is below threshold $(COVERAGE_THRESHOLD_DOMAIN)%"; \
		exit 1; \
	fi
	@# infra層のカバレッジチェック（60%以上）
	@infra_cov=$$(awk '/internal\/infra/ {gsub(/%/, "", $$NF); sum+=$$NF; count++} END {if(count>0) printf "%.1f", sum/count; else print "0"}' coverage.func.out); \
	echo "infra layer: $${infra_cov}% (threshold: $(COVERAGE_THRESHOLD_INFRA)%)"; \
	if [ $$(echo "$${infra_cov} < $(COVERAGE_THRESHOLD_INFRA)" | bc -l) -eq 1 ]; then \
		echo "ERROR: infra layer coverage $${infra_cov}% is below threshold $(COVERAGE_THRESHOLD_INFRA)%"; \
		exit 1; \
	fi
	@echo "=== All layer coverage checks passed ==="

# リリース前に実行するテスト（GoReleaserから呼び出される）
# 高速なチェックから順に実行し、早期に失敗を検出する
test-release: fmt-check lint depcheck test-coverage test-integration test-e2e

lint:
	golangci-lint run ./...

depcheck:
	go vet -vettool=$$(which depcheck) ./...

fmt:
	go fmt ./...
	go list -f '{{.Dir}}' ./... | xargs goimports -w

# フォーマット済みかどうかをチェック（CI/リリース前チェック用）
# fmtターゲットと同じくgoimportsを使用し、全パッケージを再帰的にチェック
fmt-check:
	@echo "Checking code formatting..."
	@unformatted=$$(go list -f '{{.Dir}}' ./... | xargs goimports -l); \
	if [ -n "$$unformatted" ]; then \
		echo "The following files are not formatted:"; \
		echo "$$unformatted"; \
		exit 1; \
	fi
	@echo "All files are properly formatted."

generate:
	go generate ./...
