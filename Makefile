BINARY_NAME=ai-feed
VERSION?=dev
COVERAGE_THRESHOLD_DOMAIN=80
COVERAGE_THRESHOLD_INFRA=60
COVERAGE_THRESHOLD_APP=50

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

# カバレッジレポート生成の共通ロジック
# 引数: $(1)=type (ut/it)
define generate_coverage_report
@grep -v "mock_" coverage.$(1).out > coverage.$(1).filtered.out
@mkdir -p public/coverage/$(1)
@go tool cover -html=coverage.$(1).filtered.out -o public/coverage/$(1)/index.html
@go tool cover -func=coverage.$(1).filtered.out > coverage.$(1).func.out
endef

# GitHub Actionsジョブサマリー用のカバレッジレポート生成ロジック
# 引数: なし（coverage.func.outファイルとGITHUB_STEP_SUMMARY環境変数を使用）
define generate_coverage_summary
@if [ -z "$(GITHUB_STEP_SUMMARY)" ]; then \
	echo "GITHUB_STEP_SUMMARY is not set. Skipping coverage summary (GitHub Actions only)."; \
	exit 0; \
fi
@echo "## 📊 カバレッジレポート" >> $(GITHUB_STEP_SUMMARY)
@echo "" >> $(GITHUB_STEP_SUMMARY)
@echo "### 層別カバレッジ" >> $(GITHUB_STEP_SUMMARY)
@echo "" >> $(GITHUB_STEP_SUMMARY)
@echo "| 層 | カバレッジ | 目標 | 状態 |" >> $(GITHUB_STEP_SUMMARY)
@echo "|---|---|---|---|" >> $(GITHUB_STEP_SUMMARY)
@for layer in app domain infra cmd; do \
	cov=$$(awk "/internal\/$${layer}\// {gsub(/%/, \"\", \$$NF); sum+=\$$NF; count++} END {if(count>0) printf \"%.1f\", sum/count; else print \"0\"}" coverage.func.out); \
	case $${layer} in \
		domain) threshold=$(COVERAGE_THRESHOLD_DOMAIN); target="$(COVERAGE_THRESHOLD_DOMAIN)%"; ;; \
		infra) threshold=$(COVERAGE_THRESHOLD_INFRA); target="$(COVERAGE_THRESHOLD_INFRA)%"; ;; \
		app) threshold=$(COVERAGE_THRESHOLD_APP); target="$(COVERAGE_THRESHOLD_APP)%"; ;; \
		*) threshold=0; target="-"; ;; \
	esac; \
	if [ "$${target}" = "-" ]; then \
		status="➖"; \
	elif [ $$(awk "BEGIN {print ($${cov} >= $${threshold})}") -eq 1 ]; then \
		status="✅"; \
	else \
		status="❌"; \
	fi; \
	echo "| $${layer} | $${cov}% | $${target} | $${status} |" >> $(GITHUB_STEP_SUMMARY); \
done
@echo "" >> $(GITHUB_STEP_SUMMARY)
@echo "### レポートリンク" >> $(GITHUB_STEP_SUMMARY)
@echo "" >> $(GITHUB_STEP_SUMMARY)
@echo "- [ユニットテストカバレッジ](../../../actions/artifacts) (coverage-report-ut)" >> $(GITHUB_STEP_SUMMARY)
@echo "- [統合テストカバレッジ](../../../actions/artifacts) (coverage-report-it)" >> $(GITHUB_STEP_SUMMARY)
@echo "" >> $(GITHUB_STEP_SUMMARY)
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
	rm -f coverage*.out coverage*.func.out coverage.html
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
	$(call generate_coverage_report,ut)

# 結合テストのカバレッジレポート生成
test-coverage-it:
	@go test -tags=integration -coverprofile=coverage.it.out -coverpkg=./internal/... ./test/integration/...
	$(call generate_coverage_report,it)

# ユニットテスト+結合テストの統合カバレッジレポート生成（層別カバレッジチェック含む）
test-coverage: test-coverage-ut test-coverage-it
	@echo "mode: set" > coverage.out
	@tail -n +2 coverage.ut.out >> coverage.out
	@tail -n +2 coverage.it.out >> coverage.out
	@grep -v "mock_" coverage.out > coverage.filtered.out
	@go tool cover -func=coverage.filtered.out > coverage.func.out
	@echo "=== Layer Coverage Check (per docs/03_testing_rules.md) ==="
	$(call check_layer_coverage,app,COVERAGE_THRESHOLD_APP)
	$(call check_layer_coverage,domain,COVERAGE_THRESHOLD_DOMAIN)
	$(call check_layer_coverage,infra,COVERAGE_THRESHOLD_INFRA)
	@echo "=== All layer coverage checks passed ==="

# GitHub Actionsジョブサマリー用のカバレッジレポート出力
# test-coverageの後に実行し、$GITHUB_STEP_SUMMARYに追記する
# ローカル実行時（GITHUB_STEP_SUMMARY未定義）はスキップする
coverage-summary:
	$(call generate_coverage_summary)

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
