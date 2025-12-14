BINARY_NAME=ai-feed
VERSION?=dev
COVERAGE_THRESHOLD_DOMAIN=80
COVERAGE_THRESHOLD_INFRA=60

# 層別カバレッジチェックの共通ロジック
# 引数: $(1)=層名(domain/infra), $(2)=しきい値変数名(COVERAGE_THRESHOLD_DOMAIN/COVERAGE_THRESHOLD_INFRA)
define check_layer_coverage
@layer_cov=$$(awk '/internal\/$(1)\// {gsub(/%/, "", $$NF); sum+=$$NF; count++} END {if(count>0) printf "%.1f", sum/count; else print "0"}' coverage.func.out); \
echo "$(1) layer: $${layer_cov}% (threshold: $($(2))%)"; \
if [ $$(awk "BEGIN {print ($${layer_cov} < $($(2)))}") -eq 1 ]; then \
	echo "ERROR: $(1) layer coverage $${layer_cov}% is below threshold $($(2))%"; \
	exit 1; \
fi
endef

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
	rm -f coverage.ut.out coverage.ut.filtered.out coverage.ut.func.out
	rm -f coverage.it.out coverage.it.filtered.out coverage.it.func.out
	rm -rf public/coverage

test:
	go test ./...

test-integration:
	go test -tags=integration ./test/integration/...

test-e2e:
	go test -tags=e2e -v ./test/e2e/...

# ユニットテストのカバレッジレポート生成
test-coverage-ut:
	@go test -coverprofile=coverage.ut.out -coverpkg=./internal/... ./...
	@grep -v "mock_" coverage.ut.out > coverage.ut.filtered.out
	@mkdir -p public/coverage/ut
	@go tool cover -html=coverage.ut.filtered.out -o public/coverage/ut/index.html
	@go tool cover -func=coverage.ut.filtered.out > coverage.ut.func.out

# 結合テストのカバレッジレポート生成
test-coverage-it:
	@go test -tags=integration -coverprofile=coverage.it.out -coverpkg=./internal/... ./test/integration/...
	@grep -v "mock_" coverage.it.out > coverage.it.filtered.out
	@mkdir -p public/coverage/it
	@go tool cover -html=coverage.it.filtered.out -o public/coverage/it/index.html
	@go tool cover -func=coverage.it.filtered.out > coverage.it.func.out

# ユニットテスト+結合テストの統合カバレッジレポート生成（層別カバレッジチェック含む）
test-coverage: test-coverage-ut test-coverage-it
	@go test -tags=integration -coverprofile=coverage.out -coverpkg=./internal/... ./... ./test/integration/...
	@grep -v "mock_" coverage.out > coverage.filtered.out
	@go tool cover -func=coverage.filtered.out > coverage.func.out
	@echo "=== Layer Coverage Check (per docs/03_testing_rules.md) ==="
	$(call check_layer_coverage,domain,COVERAGE_THRESHOLD_DOMAIN)
	$(call check_layer_coverage,infra,COVERAGE_THRESHOLD_INFRA)
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
