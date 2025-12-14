# テストガイド

ai-feedプロジェクトにおけるテストの書き方と実行方法について説明します。

## テストフレームワーク

| ツール | 用途 |
|--------|------|
| testify | アサーション（`assert`/`require`）とテストスイート |
| go.uber.org/mock (gomock) | モックの生成と管理 |
| 標準testingパッケージ | Goの標準テストフレームワーク |

## 層別テスト戦略

4層アーキテクチャに対応したテスト戦略を定義しています。

| 層 | 主なテスト種類 | カバレッジ目標 | 理由 |
|---|---|---|---|
| **cmd** | E2Eテスト | 設定なし | フラグ解析のみでE2Eで十分 |
| **app** | ユニット + 統合テスト | 50%以上 | 条件分岐はユニット、infra連携は統合で検証 |
| **domain** | ユニット + 統合テスト | 80%以上 | ビジネスルール最重要、連携は統合で検証 |
| **infra** | ユニット + 統合テスト | 60%以上 | 外部依存のモック化が複雑 |

### cmd層（Presentation Layer）

**テスト方針**: E2Eテストで担保し、ユニットテストは原則不要。

cmd層はフラグ解析と依存性注入のみを担当し、ビジネスロジックを含まないため、E2Eテストで十分カバー可能です。詳細は [test/e2e/README.md](../test/e2e/README.md) を参照。

### app層（Application Layer）

| テスト種類 | 用途 | 配置場所 |
|-----------|------|---------|
| ユニットテスト | 条件分岐、バリデーション、エラーハンドリング | `internal/app/*_test.go` |
| 統合テスト | infra連携、ファイル操作、並行処理 | `test/integration/app/*_test.go` |

### domain層（Domain Layer）

| テスト種類 | 用途 | 配置場所 |
|-----------|------|---------|
| ユニットテスト | 単一エンティティのロジック、バリデーション | `internal/domain/**/*_test.go` |
| 統合テスト | 複数エンティティ間のマージ・連携 | `test/integration/domain/**/*_test.go` |

**重要**: domain層の統合テストでは外部依存（ファイルI/O、DB、API）を使用しないでください。

### infra層（Infrastructure Layer）

| テスト種類 | 用途 | 配置場所 |
|-----------|------|---------|
| ユニットテスト | APIリクエスト/レスポンス処理（httptestでモック） | `internal/infra/**/*_test.go` |
| 統合テスト | 実際の外部システムとの連携 | `test/integration/infra/*_test.go` |

## テストの実行

```bash
# 全テストの実行
make test

# 統合テスト
make test-integration

# E2Eテスト
make test-e2e

# カバレッジ付きで実行
go test -cover ./...

# 特定パッケージのテスト
go test ./internal/domain/...

# 特定のテスト関数のみ
go test -run TestNewRecommendRunner ./cmd/runner/
```

### カバレッジ確認

```bash
# 層別カバレッジ
go test -cover ./internal/domain/...   # 目標: 80%以上
go test -cover ./internal/app/...      # 目標: 50%以上
go test -cover ./internal/infra/...    # 目標: 60%以上

# HTMLレポート生成
go test -coverprofile=coverage.out ./...
go tool cover -html=coverage.out
```

**GitHub Actionsでの確認方法**:
- **ジョブサマリー**: ワークフロー実行結果ページで層別カバレッジを確認
- **Artifacts**: `coverage-report-ut`（ユニット）/ `coverage-report-it`（統合）
- **GitHub Pages**: https://canpok1.github.io/ai-feed/coverage/

## テーブル駆動テスト

複数のテストケースを効率的に管理するパターンです。詳細は [01_coding_rules.md](./01_coding_rules.md#61-テーブル駆動テスト) を参照。

```go
func TestArticle_Validate(t *testing.T) {
    tests := []struct {
        name    string
        article entity.Article
        wantErr bool
    }{
        {name: "正常系: 有効な記事", article: entity.Article{Title: "Test", Link: "https://example.com"}, wantErr: false},
        {name: "異常系: リンクが空", article: entity.Article{Title: "Test", Link: ""}, wantErr: true},
    }

    for _, tt := range tests {
        t.Run(tt.name, func(t *testing.T) {
            err := tt.article.Validate()
            if tt.wantErr {
                assert.Error(t, err)
            } else {
                assert.NoError(t, err)
            }
        })
    }
}
```

## モックの使用

詳細は [01_coding_rules.md](./01_coding_rules.md#62-モックの使用) を参照。

```bash
# 全モックの再生成
make generate
```

**モック戦略**:
- **domain層インターフェース**: go.uber.org/mockで生成したモックを使用（ユニットテスト）
- **HTTP API**: httptestパッケージでモックサーバーを作成（統合テスト）

```go
// infra層の統合テストでのhttptest使用例（外部APIをモック化）
server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
    w.WriteHeader(http.StatusOK)
    json.NewEncoder(w).Encode(map[string]bool{"ok": true})
}))
defer server.Close()
```

## アサーションのベストプラクティス

| 関数 | 用途 |
|------|------|
| `assert.Equal` | 値の等価性確認（テスト継続） |
| `assert.NoError` | エラーなし確認（テスト継続） |
| `require.NoError` | エラーなし確認（失敗時即座に終了） |
| `require.NotNil` | nilでないことの確認（失敗時即座に終了） |
| `assert.ErrorIs` | Sentinel errorの確認 |
| `assert.ErrorAs` | エラー型の確認 |

## 統合テスト

### ファイル配置

```
test/integration/
├── common/                        # 共通ユーティリティ
├── app/                           # app層の統合テスト
│   └── testdata/
├── domain/                        # domain層の統合テスト
│   └── config/
└── infra/                         # infra層の統合テスト
    └── testdata/
```

### ビルドタグ

統合テストファイルには `//go:build integration` タグを追加：

```go
//go:build integration

package config

func TestConfigLoader_ValidYAML(t *testing.T) {
    // 統合テストの実装
}
```

### 実行方法

```bash
# 統合テストを実行
make test-integration

# 直接実行
go test -tags=integration -v ./test/integration/...
```

### ベストプラクティス

1. **一時ディレクトリを活用**: `t.TempDir()` を使用（自動削除）
2. **testdataディレクトリにテストデータを配置**: バージョン管理に含め再現性確保
3. **外部サービスはモックサーバーを使用**: `httptest.NewServer`
4. **テスト間の独立性を保つ**: 共有リソースは適切にクリーンアップ

## E2Eテスト

E2Eテストの詳細は [test/e2e/README.md](../test/e2e/README.md) を参照してください。

### 概要

- 実際のバイナリをビルドして実行
- エンドユーザーの視点で動作を検証
- `//go:build e2e` タグを使用

### 実行方法

```bash
# E2Eテストを実行
make test-e2e

# 直接実行
go test -tags=e2e -v ./test/e2e/...
```

## テストのチェックリスト

新しいテストを書く際の確認事項：

- [ ] テーブル駆動テストパターンを使用
- [ ] 正常系と異常系の両方をカバー
- [ ] モックを適切に使用
- [ ] テスト名が内容を明確に表している
- [ ] 必要に応じて `t.Parallel()` を使用
- [ ] クリーンアップは `defer` または `t.Cleanup()` を使用

## セキュリティ考慮事項

テスト実装時に考慮すべきセキュリティ対策：

- APIキーは環境変数または設定ファイルで管理（`.gitignore`で除外）
- URLの妥当性チェック
- プロンプトインジェクション対策
- サニタイゼーション処理

## トラブルシューティング

| 問題 | 解決方法 |
|------|----------|
| モックのエラー | `make generate` でモックを再生成 |
| 依存関係の問題 | `go mod tidy && go mod download` |
| テストキャッシュ | `go clean -testcache` |
| 詳細ログが必要 | `go test -v -run TestName ./path/to/package` |
