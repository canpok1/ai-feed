# テストガイド

ai-feedプロジェクトにおけるテストの書き方と実行方法について説明します。

## テストフレームワーク

### 使用しているツール

- **testify**: アサーションとテストスイート
  - `assert`: テストのアサーション
  - `require`: 失敗時にテストを即座に終了
- **go.uber.org/mock (gomock)**: モックの生成と管理
- **標準testingパッケージ**: Goの標準テストフレームワーク

## テストの実行

### 基本的なテスト実行

```bash
# 全テストの実行
make test

# 特定のパッケージのテスト実行
go test ./internal/domain/...
go test ./cmd/runner/...

# 詳細な出力でテスト実行
go test -v ./...

# 特定のテスト関数のみ実行
go test -run TestNewRecommendRunner ./cmd/runner/

# 並列実行数を指定してテスト
go test -parallel 4 ./...
```

### テストカバレッジ

```bash
# カバレッジ付きでテスト実行
go test -cover ./...

# カバレッジレポートの生成
go test -coverprofile=coverage.out ./...

# HTMLレポートの表示
go tool cover -html=coverage.out

# パッケージごとのカバレッジ確認
go test -coverprofile=coverage.out -covermode=atomic ./...
go tool cover -func=coverage.out
```

## テーブル駆動テスト

ai-feedでは、テーブル駆動テストパターンを採用しています。これにより、複数のテストケースを効率的に管理できます。

### 基本的な構造

```go
func TestFunctionName(t *testing.T) {
    tests := []struct {
        name     string        // テストケース名
        input    interface{}   // 入力値
        want     interface{}   // 期待値
        wantErr  bool         // エラー期待フラグ
    }{
        {
            name:    "正常系: 基本的なケース",
            input:   "test",
            want:    "expected",
            wantErr: false,
        },
        {
            name:    "異常系: 無効な入力",
            input:   "",
            want:    nil,
            wantErr: true,
        },
    }

    for _, tt := range tests {
        t.Run(tt.name, func(t *testing.T) {
            // テストの実装
            got, err := FunctionToTest(tt.input)
            
            if tt.wantErr {
                assert.Error(t, err)
                return
            }
            
            require.NoError(t, err)
            assert.Equal(t, tt.want, got)
        })
    }
}
```

### 実際の例（cmd/runner/recommend_test.go より）

```go
func TestNewRecommendRunner(t *testing.T) {
    tests := []struct {
        name             string
        outputConfig     *infra.OutputConfig
        promptConfig     *infra.PromptConfig
        expectError      bool
        expectedErrorMsg string
    }{
        {
            name:         "Successful creation with no viewers",
            outputConfig: &infra.OutputConfig{},
            promptConfig: &infra.PromptConfig{CommentPromptTemplate: "test-template"},
            expectError:  false,
        },
        {
            name: "Successful creation with SlackAPI viewer",
            outputConfig: &infra.OutputConfig{
                SlackAPI: &infra.SlackAPIConfig{
                    APIToken:        "test-token",
                    Channel:         "#test",
                    MessageTemplate: stringPtr("{{.Article.Title}}\n{{.Article.Link}}"),
                },
            },
            promptConfig: &infra.PromptConfig{CommentPromptTemplate: "test-template"},
            expectError:  false,
        },
    }

    for _, tt := range tests {
        t.Run(tt.name, func(t *testing.T) {
            ctrl := gomock.NewController(t)
            defer ctrl.Finish()

            mockFetchClient := mock_domain.NewMockFetchClient(ctrl)
            mockRecommender := mock_domain.NewMockRecommender(ctrl)

            runner, err := NewRecommendRunner(
                mockFetchClient,
                mockRecommender,
                tt.outputConfig,
                tt.promptConfig,
            )

            if tt.expectError {
                assert.Error(t, err)
                assert.Contains(t, err.Error(), tt.expectedErrorMsg)
            } else {
                assert.NoError(t, err)
                assert.NotNil(t, runner)
            }
        })
    }
}
```

## モックの使用

### モックの生成

```bash
# 全モックの再生成
make generate

# 特定のインターフェースのモック生成
mockgen -source=internal/domain/fetch.go -destination=internal/domain/mock_domain/fetch.go -package=mock_domain
```

### モック生成の設定（generate.go）

各パッケージに`generate.go`ファイルを配置し、モック生成を定義します：

```go
//go:generate mockgen -source=../fetch.go -destination=fetch.go -package=mock_domain
//go:generate mockgen -source=../comment.go -destination=comment.go -package=mock_domain
//go:generate mockgen -source=../message.go -destination=message.go -package=mock_domain
```

### モックの使用例

```go
func TestRecommendRunner_Run(t *testing.T) {
    ctrl := gomock.NewController(t)
    defer ctrl.Finish()

    // モックの作成
    mockFetcher := mock_domain.NewMockFetcher(ctrl)
    mockRecommender := mock_domain.NewMockRecommender(ctrl)
    mockViewer := mock_domain.NewMockViewer(ctrl)

    // 期待値の設定
    mockFetcher.EXPECT().
        Fetch(gomock.Any(), []string{"https://example.com/feed"}).
        Return([]entity.Article{
            {Title: "Test Article", Link: "https://example.com/1"},
        }, nil)

    mockRecommender.EXPECT().
        Recommend(gomock.Any(), gomock.Any(), gomock.Any(), gomock.Any()).
        Return(&entity.Recommend{
            Article: entity.Article{Title: "Test Article"},
            Comment: "Test comment",
        }, nil)

    mockViewer.EXPECT().
        SendRecommend(gomock.Any(), gomock.Any()).
        Return(nil)

    // テスト実行
    runner := &RecommendRunner{
        fetcher:     mockFetcher,
        recommender: mockRecommender,
        viewers:     []domain.Viewer{mockViewer},
    }

    err := runner.Run(context.Background(), params, profile)
    assert.NoError(t, err)
}
```

## テストヘルパー関数

### 共通のテストユーティリティ

```go
// ポインタ変換ヘルパー
func stringPtr(s string) *string {
    return &s
}

func intPtr(i int) *int {
    return &i
}

// テスト用の設定作成
func createTestConfig() *infra.Config {
    return &infra.Config{
        DefaultProfile: &infra.Profile{
            AI: &infra.AIConfig{
                Gemini: &infra.GeminiConfig{
                    Type:   "test-type",
                    APIKey: "test-key",
                },
            },
        },
    }
}
```

### 一時ファイル・ディレクトリの使用

```go
func TestFileOperation(t *testing.T) {
    // t.TempDir()を使用して一時ディレクトリを作成
    tmpDir := t.TempDir()
    
    // テスト終了時に自動的にクリーンアップされる
    configPath := filepath.Join(tmpDir, "test.yaml")
    
    err := WriteConfig(configPath, testConfig)
    assert.NoError(t, err)
}
```

## アサーションのベストプラクティス

### assert vs require

```go
// assert: テストを継続
assert.Equal(t, expected, actual)
assert.NoError(t, err)
assert.Contains(t, message, "expected text")

// require: 失敗時にテストを即座に終了
require.NoError(t, err)  // エラーがあったら以降の検証は無意味
require.NotNil(t, result)  // nilチェック後に使用
```

### エラーメッセージの検証

```go
// エラーの存在確認
assert.Error(t, err)

// エラーメッセージの部分一致
assert.Contains(t, err.Error(), "expected error")

// Sentinel errorの確認
assert.ErrorIs(t, err, ErrNoArticlesFound)

// エラー型の確認
var customErr *CustomError
assert.ErrorAs(t, err, &customErr)
```

## 統合テスト

### 実際の外部サービスを使用するテスト

```go
func TestIntegration_RealAPI(t *testing.T) {
    // 統合テストはスキップ可能にする
    if testing.Short() {
        t.Skip("Skipping integration test in short mode")
    }

    // 環境変数のチェック
    apiKey := os.Getenv("TEST_API_KEY")
    if apiKey == "" {
        t.Skip("TEST_API_KEY not set")
    }

    // 実際のAPIを使用したテスト
}
```

### ビルドタグを使用した統合テスト

```go
//go:build integration
// +build integration

package integration_test

func TestRealAPIIntegration(t *testing.T) {
    // 統合テストの実装
}
```

実行方法：
```bash
# 統合テストを実行
go test -tags=integration ./...
```

## ベンチマークテスト

### ベンチマークの書き方

```go
func BenchmarkFetch(b *testing.B) {
    fetcher := NewFetcher()
    
    b.ResetTimer()
    for i := 0; i < b.N; i++ {
        fetcher.Fetch(context.Background(), testURLs)
    }
}

// サブベンチマーク
func BenchmarkParser(b *testing.B) {
    b.Run("RSS", func(b *testing.B) {
        for i := 0; i < b.N; i++ {
            parseRSS(testRSSData)
        }
    })
    
    b.Run("Atom", func(b *testing.B) {
        for i := 0; i < b.N; i++ {
            parseAtom(testAtomData)
        }
    })
}
```

実行方法：
```bash
# ベンチマーク実行
go test -bench=.

# 特定のベンチマークのみ
go test -bench=BenchmarkFetch

# メモリアロケーションも計測
go test -bench=. -benchmem
```

## テストのデバッグ

### デバッグ出力

```go
func TestDebug(t *testing.T) {
    // t.Logはテスト失敗時か-vオプション時のみ出力
    t.Log("Debug information:", value)
    
    // t.Logfでフォーマット指定
    t.Logf("Processing item %d: %v", index, item)
}
```

### テストの並列実行

```go
func TestParallel(t *testing.T) {
    t.Parallel()  // このテストを並列実行可能にする
    
    tests := []struct{
        name string
    }{
        {"case1"},
        {"case2"},
    }
    
    for _, tt := range tests {
        tt := tt  // ループ変数のキャプチャ
        t.Run(tt.name, func(t *testing.T) {
            t.Parallel()  // サブテストも並列実行
            // テスト実装
        })
    }
}
```

## CI/CDでのテスト

### GitHub Actionsでの設定例

```yaml
name: Test
on: [push, pull_request]

jobs:
  test:
    runs-on: ubuntu-latest
    steps:
      - uses: actions/checkout@v4
      - uses: actions/setup-go@v5
        with:
          go-version: '1.24'
      
      - name: Run tests
        run: |
          make test
          
      - name: Upload coverage
        uses: codecov/codecov-action@v3
        with:
          file: ./coverage.out
```

## テストのチェックリスト

新しいテストを書く際は、以下を確認してください：

- [ ] テーブル駆動テストパターンを使用している
- [ ] 正常系と異常系の両方をカバーしている
- [ ] モックを適切に使用している
- [ ] エラーケースを網羅している
- [ ] テスト名が内容を明確に表している
- [ ] 必要に応じてt.Parallel()を使用している
- [ ] クリーンアップが必要な場合はdeferまたはt.Cleanup()を使用している
- [ ] アサーションメッセージが分かりやすい

## トラブルシューティング

### テストが失敗する場合

1. モックの再生成
   ```bash
   make generate
   ```

2. 依存関係の更新
   ```bash
   go mod tidy
   go mod download
   ```

3. テストキャッシュのクリア
   ```bash
   go clean -testcache
   ```

4. 詳細ログで実行
   ```bash
   go test -v -run TestName ./path/to/package
   ```

### モックのエラー

- `Unexpected call`エラー: EXPECT()の設定を確認
- `Missing call`エラー: 期待された呼び出しが実行されていない
- タイムアウト: コンテキストのキャンセルやタイムアウト設定を確認