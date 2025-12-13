# テストガイド

ai-feedプロジェクトにおけるテストの書き方と実行方法について説明します。

## テストフレームワーク

### 使用しているツール

- **testify**: アサーションとテストスイート
  - `assert`: テストのアサーション
  - `require`: 失敗時にテストを即座に終了
- **go.uber.org/mock (gomock)**: モックの生成と管理
- **標準testingパッケージ**: Goの標準テストフレームワーク

## 層別テスト戦略

4層アーキテクチャに対応したテスト戦略を定義しています。各層の特性に応じて、適切なテスト種類とカバレッジ目標を設定します。

### テスト戦略の概要

| 層 | 主なテスト種類 | カバレッジ目標 | 理由 |
|---|---|---|---|
| **cmd (Presentation)** | E2Eテスト | 設定なし | フラグ解析のみで、E2Eで十分に担保可能 |
| **app (Application)** | 統合テスト | 設定なし | オーケストレーションは実際のコンポーネント連携で検証すべき |
| **domain (Domain)** | ユニットテスト | 80%以上 | ビジネスルールの正確性が最重要、純粋なロジックでテストしやすい |
| **infra (Infrastructure)** | ユニット + 統合テスト | 60%以上 | 外部依存のモック化が複雑なため |

### 1. cmd層（Presentation Layer）のテスト

**方針**: E2Eテストで担保し、ユニットテストは最小限

cmd層はフラグ解析とApplication層への委譲のみを行うため、個別のユニットテストよりもE2Eテストでの検証が効果的です。

```go
// cmd層のテストは基本的に不要
// E2Eテスト（test/e2e/）で以下を検証：
// - コマンドライン引数の解析
// - エラーメッセージの表示
// - 終了コード
```

**テスト対象**:
- ❌ ユニットテスト: 原則不要（カバレッジ目標なし）
- ✅ E2Eテスト: コマンド全体の動作を検証

### 2. app層（Application Layer）のテスト

**方針**: 統合テストでオーケストレーションを検証

app層は複数コンポーネントの協調動作（オーケストレーション）を担うため、モックではなく実際のinfra層実装と組み合わせた統合テストで検証します。

```go
//go:build integration

// test/integration/app/recommend_test.go
// ※簡潔さのため、import文やセットアップコード（config, ctx, params等）は省略しています
func TestRecommendUseCase_Execute(t *testing.T) {
    // 実際のinfra層実装を使用
    fetcher := fetch.NewRSSFetcher(http.DefaultClient)

    // 外部APIのみモックサーバーを使用
    slackHandler := http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
        w.WriteHeader(http.StatusOK)
    })
    mockSlackServer := httptest.NewServer(slackHandler)
    defer mockSlackServer.Close()

    slackSender := message.NewSlackSender(config, slack.New("token", slack.OptionAPIURL(mockSlackServer.URL+"/")))

    // 実際のコンポーネントでUseCaseを構築
    useCase := app.NewRecommendUseCase(fetcher, recommender, []domain.MessageSender{slackSender})

    // 統合テスト実行
    err := useCase.Execute(ctx, params)
    assert.NoError(t, err)
}
```

**テスト対象**:
- ✅ 複数コンポーネントの協調動作
- ✅ 設定の読み込み・マージ・検証の一連のフロー
- ✅ エラーハンドリング（実際のエラー伝播）

**ユニットテストについて**:
- ❌ オーケストレーションのモックテストは原則不要
- ⚠️ 複雑な条件分岐がある場合のみ、個別にユニットテストを追加

### 3. domain層（Domain Layer）のテスト

**方針**: 高カバレッジのユニットテストでビジネスルールを厳密に検証

domain層は外部依存がなく純粋なロジックのため、テストが容易です。ビジネスルールの正確性を保証するため、最も高いカバレッジを維持します。

```go
// internal/domain/entity/article_test.go
func TestArticle_Validate(t *testing.T) {
    tests := []struct {
        name    string
        article entity.Article
        wantErr bool
    }{
        {
            name:    "正常系: 有効な記事",
            article: entity.Article{Title: "Test", Link: "https://example.com"},
            wantErr: false,
        },
        {
            name:    "異常系: リンクが空",
            article: entity.Article{Title: "Test", Link: ""},
            wantErr: true,
        },
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

**テスト対象**:
- ✅ エンティティのバリデーション
- ✅ 値オブジェクトの等価性・変換
- ✅ ドメインサービスのロジック
- ✅ 境界値・エッジケース

### 4. infra層（Infrastructure Layer）のテスト

**方針**: ユニットテスト + 統合テストの組み合わせ

infra層は外部システムとの連携を担うため、モックを使用したユニットテストと、実際の外部システムを使用した統合テストを組み合わせます。

```go
// internal/infra/message/slack_test.go（ユニットテスト）
func TestSlackSender_SendRecommend(t *testing.T) {
    // httptest.NewServerでSlack APIをモック
    server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
        w.WriteHeader(http.StatusOK)
        json.NewEncoder(w).Encode(map[string]bool{"ok": true})
    }))
    defer server.Close()

    sender := NewSlackSender(config, slack.New("token", slack.OptionAPIURL(server.URL+"/")))
    err := sender.SendRecommend(recommend, "message")
    assert.NoError(t, err)
}
```

**テスト対象**:
- ✅ domain層インターフェースの実装
- ✅ 外部APIのリクエスト/レスポンス処理
- ✅ ファイルI/O処理
- ✅ エラーハンドリング（ネットワークエラー等）

### カバレッジ確認コマンド

```bash
# 層別カバレッジの確認
go test -cover ./internal/domain/...   # 目標: 80%以上
go test -cover ./internal/app/...      # 統合テストで検証（カバレッジ目標なし）
go test -cover ./internal/infra/...    # 目標: 60%以上
go test -cover ./cmd/...               # 目標: 設定なし

# 全体カバレッジレポート
go test -coverprofile=coverage.out ./...
go tool cover -func=coverage.out
```

## テストの実行

### 基本的なテスト実行

```bash
# 全テストの実行
make test

# 特定のパッケージのテスト実行
go test ./internal/domain/...
go test ./internal/app/...

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

### 実際の例（internal/app/recommend_test.go より）

```go
func TestNewRecommendUseCase(t *testing.T) {
    tests := []struct {
        name             string
        outputConfig     *entity.OutputConfig
        promptConfig     *entity.PromptConfig
        expectError      bool
        expectedErrorMsg string
    }{
        {
            name:         "正常系: senderなしで作成",
            outputConfig: &entity.OutputConfig{},
            promptConfig: &entity.PromptConfig{CommentPromptTemplate: "test-template"},
            expectError:  false,
        },
        {
            name: "正常系: SlackAPI senderありで作成",
            outputConfig: &entity.OutputConfig{
                SlackAPI: &entity.SlackAPIConfig{
                    APIToken:        entity.NewSecret("test-token"),
                    Channel:         "#test",
                    MessageTemplate: stringPtr("{{.Article.Title}}\n{{.Article.Link}}"),
                },
            },
            promptConfig: &entity.PromptConfig{CommentPromptTemplate: "test-template"},
            expectError:  false,
        },
    }

    for _, tt := range tests {
        t.Run(tt.name, func(t *testing.T) {
            ctrl := gomock.NewController(t)
            defer ctrl.Finish()

            mockFetchClient := mock_domain.NewMockFetchClient(ctrl)
            mockRecommender := mock_domain.NewMockRecommender(ctrl)

            useCase, err := NewRecommendUseCase(
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
                assert.NotNil(t, useCase)
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

モックは主にdomain層やinfra層のユニットテストで使用します。
**注意**: app層のテストでは統合テストを推奨しているため、モックの使用は限定的です。

```go
// infra層のテスト例（internal/infra/message/slack_test.go）
func TestSlackSender_SendRecommend(t *testing.T) {
    // 外部APIのモックサーバーを作成
    server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
        w.WriteHeader(http.StatusOK)
        json.NewEncoder(w).Encode(map[string]bool{"ok": true})
    }))
    defer server.Close()

    // テスト対象のインスタンスを作成
    sender := NewSlackSender(config, slack.New("token", slack.OptionAPIURL(server.URL+"/")))

    // テスト実行
    recommend := &entity.Recommend{
        Article: entity.Article{Title: "Test Article", Link: "https://example.com/1"},
        Comment: stringPtr("Test comment"),
    }
    err := sender.SendRecommend(recommend, "message")
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

## 統合テスト (Integration Test)

### 統合テストの概要

統合テストは、複数のコンポーネント間の連携を検証するテストです。ユニットテストでは個別のコンポーネントをモックを使用して分離してテストしますが、統合テストでは実際のコンポーネントを組み合わせて、正しく連携できることを確認します。

### 目的と対象範囲

統合テストの主な目的は以下のとおりです：

1. **複数コンポーネント間の連携検証**
   - リポジトリ層とドメイン層の連携
   - 外部APIクライアントとドメインロジックの連携
   - 設定ファイルの読み込みから処理までの一連のフロー

2. **境界条件の検証**
   - ユニットテストでモック化していた外部依存を実際に使用
   - 実際のファイル操作やネットワーク呼び出しを伴うシナリオ

3. **設定ファイル処理の検証**
   - YAML設定ファイルの読み込み・パース・バリデーション
   - プロファイル設定の継承・マージ処理

### テストの種類と違い

ai-feedプロジェクトでは3種類のテストを使い分けています：

| 項目 | ユニットテスト | 統合テスト | E2Eテスト |
|------|---------------|-----------|----------|
| **対象範囲** | 単一の関数・メソッド | 複数コンポーネントの連携 | アプリケーション全体 |
| **モック使用** | 依存を全てモック化 | 外部サービス以外は実装を使用 | 外部APIのみモック化 |
| **実行速度** | 高速 | 比較的低速 | 低速 |
| **配置場所** | `<package>/*_test.go` | `test/integration/**/*_test.go` | `test/e2e/**/*_test.go` |
| **ビルドタグ** | なし | `integration` | `e2e` |
| **実行コマンド** | `make test` | `make test-integration` | `make test-e2e` |
| **目的** | ロジックの正確性 | コンポーネント間の連携 | 実際のユーザー操作の再現 |
| **主な対象層** | domain, infra | app, infra | cmd（Presentation層全体） |

### ファイル配置ルール

統合テストは `test/integration/` ディレクトリに配置します。

```
test/integration/
├── common/                        # 共通ユーティリティ
│   └── helper.go                  # テストヘルパー関数
├── config/                        # 設定ファイル処理のテスト
│   ├── config_test.go             # 設定読み込みの統合テスト
│   ├── main_test.go               # パッケージセットアップ
│   └── testdata/                  # テスト用設定ファイル
│       ├── valid_config.yaml
│       └── invalid_config.yaml
└── infra/                         # インフラ層のテスト
    ├── repository_test.go         # リポジトリの統合テスト
    ├── main_test.go               # パッケージセットアップ
    └── testdata/                  # テスト用データ
        └── ...
```

#### ファイル命名規則

- テストファイル: `*_test.go`
- テスト関数: `Test<コンポーネント>_<シナリオ>`（例: `TestConfigLoader_ValidYAML`）
- テストデータ: `testdata/` ディレクトリに配置

### 統合テストの実行方法

```bash
# 統合テストを実行
make test-integration

# または直接実行
go test -tags=integration -v ./test/integration/...

# 特定のテストのみ実行
go test -tags=integration -v -run TestConfigLoader ./test/integration/config/
```

通常の `go test ./...` では実行されず、`-tags=integration` を指定した時のみ実行されます。

### ビルドタグの使用

統合テストファイルには `//go:build integration` タグを追加します：

```go
//go:build integration

package config

import (
    "testing"

    "github.com/stretchr/testify/assert"
)

func TestConfigLoader_ValidYAML(t *testing.T) {
    // 統合テストの実装
}
```

### 統合テストの書き方

#### 基本的な構造

```go
//go:build integration

package config

import (
    "os"
    "path/filepath"
    "testing"

    "github.com/stretchr/testify/assert"
    "github.com/stretchr/testify/require"
    "github.com/canpok1/ai-feed/internal/infra"
)

func TestConfigLoader_LoadValidConfig(t *testing.T) {
    // テストデータのパスを取得
    testdataDir := filepath.Join("testdata")
    configPath := filepath.Join(testdataDir, "valid_config.yaml")

    // 実際の設定ファイルを読み込む
    repo := infra.NewYamlConfigRepository(configPath)
    config, err := repo.Load()

    // 結果を検証
    require.NoError(t, err)
    assert.NotNil(t, config)
    assert.Equal(t, "expected-value", config.DefaultProfile.AI.Gemini.Type)
}

func TestConfigLoader_InvalidConfig(t *testing.T) {
    testdataDir := filepath.Join("testdata")
    configPath := filepath.Join(testdataDir, "invalid_config.yaml")

    repo := infra.NewYamlConfigRepository(configPath)
    _, err := repo.Load()

    assert.Error(t, err)
    assert.Contains(t, err.Error(), "validation failed")
}
```

#### テストデータの作成

テスト用データは `testdata/` ディレクトリに配置し、テストケースごとに適切なファイルを用意します：

```go
// テストデータの作成ヘルパー
func createTestConfig(t *testing.T, content string) string {
    t.Helper()

    tmpDir := t.TempDir()
    configPath := filepath.Join(tmpDir, "config.yaml")

    err := os.WriteFile(configPath, []byte(content), 0644)
    require.NoError(t, err)

    return configPath
}

func TestConfigLoader_DynamicConfig(t *testing.T) {
    configContent := `
default_profile:
  ai:
    gemini:
      type: gemini-2.0-flash
`
    configPath := createTestConfig(t, configContent)

    repo := infra.NewYamlConfigRepository(configPath)
    config, err := repo.Load()

    require.NoError(t, err)
    assert.Equal(t, "gemini-2.0-flash", config.DefaultProfile.AI.Gemini.Type)
}
```

### 統合テストのベストプラクティス

1. **一時ディレクトリを活用する**
   ```go
   tmpDir := t.TempDir()  // テスト終了時に自動削除される
   ```

2. **テストデータは testdata ディレクトリに配置する**
   - バージョン管理に含め、テストの再現性を確保
   - 動的に生成する場合は `t.TempDir()` を使用

3. **外部サービスはモックサーバーを使用する**
   - ネットワーク依存は `httptest.NewServer` でモック化
   - データベースはインメモリDBやテストコンテナを検討

4. **テスト間の独立性を保つ**
   - 各テストは他のテストに依存しない
   - 共有リソースを使用する場合は適切にクリーンアップ

5. **テーブル駆動テストを使用する**
   ```go
   tests := []struct {
       name       string
       configFile string
       wantErr    bool
   }{
       {name: "正常系", configFile: "valid_config.yaml", wantErr: false},
       {name: "異常系", configFile: "invalid_config.yaml", wantErr: true},
   }
   ```

### トラブルシューティング

#### 統合テストが通常のテストで実行される

- ファイル先頭に `//go:build integration` タグがあることを確認
- `go test` ではなく `go test -tags=integration` で実行

#### テストデータが見つからない

- `testdata/` ディレクトリへの相対パスを確認
- `go test` はパッケージディレクトリで実行されるため、相対パスはパッケージからの相対

#### テスト間で状態が共有される

- グローバル変数の使用を避ける
- 各テストで新しいインスタンスを作成する
- `t.Cleanup()` でリソースを確実に解放する

## E2Eテスト (End-to-End Testing)

### E2Eテストの概要

E2Eテストは、実際のバイナリを実行してエンドユーザーの視点で動作を検証するテストです。

各テストの種類の違いについては、[テストの種類と違い](#テストの種類と違い)を参照してください。

### E2Eテストの実行方法

```bash
# E2Eテストを実行
make test-e2e

# または直接実行
go test -tags=e2e -v ./test/e2e/...

# 特定のテストのみ実行
go test -tags=e2e -v -run TestInitCommand ./test/e2e/
```

### ビルドタグの使用

E2Eテストファイルには`//go:build e2e`タグを追加します：

```go
//go:build e2e

package e2e

import (
    "testing"
)

func TestInitCommand_CreateConfigFile(t *testing.T) {
    // E2Eテストの実装
}
```

このタグにより、通常の`go test`では実行されず、`-tags=e2e`を指定した時のみ実行されます。

### E2Eテストの書き方

E2Eテストでは、共通ヘルパー関数を使用してバイナリをビルド・実行します：

```go
//go:build e2e

package e2e

import (
    "testing"
    "github.com/stretchr/testify/assert"
)

func TestInitCommand_CreateConfigFile(t *testing.T) {
    // バイナリをビルド
    binaryPath := BuildBinary(t)

    // 一時ディレクトリを作成
    tmpDir := t.TempDir()

    // 一時ディレクトリに移動
    os.Chdir(tmpDir)

    // コマンドを実行
    output, err := ExecuteCommand(t, binaryPath, "init")

    // 結果を検証
    assert.NoError(t, err)
    assert.Contains(t, output, "config.yml を生成しました")

    // ファイルが作成されたことを確認
    _, err = os.Stat("config.yml")
    assert.NoError(t, err)
}
```

### E2Eヘルパー関数

`test/e2e/helper.go`には以下のヘルパー関数が定義されています：

#### GetProjectRoot

プロジェクトのルートディレクトリパスを取得します。

```go
projectRoot := GetProjectRoot(t)
```

#### BuildBinary

テスト用のバイナリをビルドし、一時ディレクトリに配置します。

```go
binaryPath := BuildBinary(t)
```

#### ExecuteCommand

バイナリを実行し、標準出力と標準エラー出力を結合して返します。

```go
output, err := ExecuteCommand(t, binaryPath, "init")
```

#### CreateTestConfig

テスト用の設定ファイルを作成します。

```go
configPath := CreateTestConfig(t, tmpDir, TestConfigParams{
    DefaultProfile: map[string]interface{}{
        "ai": map[string]interface{}{
            "gemini": map[string]interface{}{
                "type": "gemini-2.5-flash",
            },
        },
    },
})
```

### E2Eテストのディレクトリ構造

E2Eテストは**コマンドごとのサブディレクトリ**に分割された構造を採用しています。

```
test/e2e/
├── common/                        # 共通ユーティリティ
│   ├── helper.go                  # 共通ヘルパー関数
│   └── mock/                      # モックサーバー実装
│       ├── rss.go                 # RSS/Atomフィードモックサーバー
│       ├── slack.go               # Slackモックサーバー
│       └── misskey.go             # Misskeyモックサーバー
├── config/                        # configコマンドのテスト
│   ├── config_test.go             # configコマンドのE2Eテスト
│   ├── main_test.go               # パッケージセットアップ
│   └── testdata/                  # テスト用データ
│       └── ...
├── init/                          # initコマンドのテスト
│   ├── init_test.go               # initコマンドのE2Eテスト
│   ├── main_test.go               # パッケージセットアップ
│   └── testdata/                  # テスト用データ（必要に応じて）
├── profile/                       # profileコマンドのテスト
│   ├── profile_test.go            # profileコマンドのE2Eテスト
│   ├── main_test.go               # パッケージセットアップ
│   └── testdata/                  # テスト用データ
│       └── ...
└── recommend/                     # recommendコマンドのテスト
    ├── recommend_test.go          # recommendコマンドのE2Eテスト
    ├── main_test.go               # パッケージセットアップ
    └── testdata/                  # テスト用データ（必要に応じて）
```

#### コマンドベースのサブディレクトリ設計

このディレクトリ構造には以下の設計上の理由があります：

1. **パッケージの独立性**: 各コマンドのテストを独立したGoパッケージとして分離することで、テスト間の依存関係を最小化
2. **テストデータの局所化**: 関連するテストデータ（`testdata/`）を各コマンドのテストディレクトリにまとめ、管理しやすさを向上
3. **並列実行の効率化**: テストの並列実行時にパッケージ間の独立性を確保し、安定したテスト実行を実現
4. **コードの見通し**: コマンドごとにテストファイルを分割することで、該当するテストを見つけやすくなる

#### main_test.goの役割

各コマンドのテストディレクトリには`main_test.go`が含まれています。このファイルは`TestMain`関数を定義し、以下の役割を担います：

```go
func TestMain(m *testing.M) {
    common.SetupPackage()   // テスト実行前のセットアップ（バイナリビルドなど）
    code := m.Run()         // テスト実行
    common.CleanupPackage() // テスト終了後のクリーンアップ
    os.Exit(code)
}
```

- **バイナリの一度だけのビルド**: `sync.Once`を使用して、複数のテストパッケージが同時に実行されても一度だけビルドを行う
- **リソースの共有**: ビルドされたバイナリパスを全テストで共有
- **クリーンアップ**: テスト終了時に一時ファイルやリソースを削除

### AIモック機能

E2Eテストでは、外部AI API（Gemini）への依存を排除するためのモック機能を提供しています。これにより、APIキーなしでテストを実行できます。

#### モック機能の利点

- **CI/CD環境での実行**: 外部APIキーなしでテストを実行可能
- **テスト速度の向上**: 実際のAPI呼び出しがないため高速
- **安定したテスト結果**: 外部サービスの状態に依存しない
- **コスト削減**: API利用料金が発生しない

#### 設定オプション

| 設定項目 | 説明 | デフォルト値 |
|----------|------|--------------|
| `ai.mock.enabled` | モック機能の有効/無効 | `false` |
| `ai.mock.selector_mode` | 記事選択モード（`first`, `random`, `last`） | `first` |
| `ai.mock.comment` | モックが返す固定コメント | 空文字列 |

#### selector_modeの動作

- `first`: 記事リストの最初の記事を選択
- `random`: 記事リストからランダムに選択
- `last`: 記事リストの最後の記事を選択

#### 設定例

```yaml
default_profile:
  ai:
    mock:
      enabled: true
      selector_mode: first
      comment: "これはテスト用のモックコメントです。"
  # Gemini設定は不要（モック有効時は無視される）
```

#### E2Eテストでの使用

E2Eテストのヘルパー関数`CreateRecommendTestConfig`ではデフォルトでモックAIが有効化されます：

```go
// モックAIを使用（デフォルト）
config := common.CreateRecommendTestConfig(t, tmpDir, common.RecommendConfigParams{
    FeedURLs:        []string{feedURL},
    SlackWebhookURL: slackURL,
})

// 実際のGemini APIを使用する場合
useMockAI := false
config := common.CreateRecommendTestConfig(t, tmpDir, common.RecommendConfigParams{
    UseMockAI:       &useMockAI,
    GeminiAPIKey:    os.Getenv("GEMINI_API_KEY"),
    FeedURLs:        []string{feedURL},
    SlackWebhookURL: slackURL,
})
```

### recommendコマンドのE2Eテスト

recommendコマンドは実際のGemini APIを使用するため、特別な設定が必要です。

#### 環境変数の設定

```bash
# Gemini APIキーを設定（必須）
export GEMINI_API_KEY=your_gemini_api_key_here

# E2Eテストを実行
make test-e2e
```

**重要**: `GEMINI_API_KEY`環境変数が設定されていない場合、エラーメッセージが表示されテストは実行されません。

#### モックサーバーの使用

recommendコマンドのE2Eテストでは、以下のモックサーバーを使用します：

1. **RSSフィードモックサーバー** (`test/e2e/common/mock/rss.go`)
   - `NewMockRSSHandler()`: 標準的なRSSフィードを返す
   - `NewMockAtomHandler()`: Atomフィードを返す
   - `NewMockEmptyFeedHandler()`: 空のフィードを返す
   - `NewMockInvalidFeedHandler()`: 不正なXMLを返す
   - `NewMockErrorHandler(statusCode)`: 指定したHTTPエラーを返す

2. **Slackモックサーバー** (`test/e2e/common/mock/slack.go`)
   - Slack Webhookへのメッセージ送信を受信・記録
   - `ReceivedMessage()`: メッセージ受信確認
   - `GetMessages()`: 受信したメッセージ一覧取得
   - `Reset()`: 状態リセット

3. **Misskeyモックサーバー** (`test/e2e/common/mock/misskey.go`)
   - Misskeyのノート作成APIを模倣
   - `ReceivedNote()`: ノート受信確認
   - `GetNotes()`: 受信したノート一覧取得
   - `Reset()`: 状態リセット

#### テストケース

- `TestRecommendCommand_WithRealGeminiAPI`: 実際のGemini APIを使用した正常系テスト
- `TestRecommendCommand_WithMisskey`: Misskeyへの出力テスト
- `TestRecommendCommand_MultipleOutputs`: 複数出力先(Slack+Misskey)テスト
- `TestRecommendCommand_EmptyFeed`: 空フィードの処理テスト
- `TestRecommendCommand_InvalidFeed`: 不正なフィードの処理テスト
- `TestRecommendCommand_WithProfile`: プロファイル使用時のテスト

#### GitHub ActionsでのE2Eテスト実行

リポジトリのSecretsに`GEMINI_API_KEY`を設定することで、CIでもE2Eテストが実行されます：

1. GitHub リポジトリ > Settings > Secrets and variables > Actions
2. "New repository secret" をクリック
3. Name: `GEMINI_API_KEY`、Secret: Gemini APIキーを入力
4. "Add secret" をクリック

設定後、プルリクエストやmainブランチへのプッシュ時に自動的にE2Eテストが実行されます。

### E2Eテストのベストプラクティス

1. **一時ディレクトリを使用する**
   ```go
   tmpDir := t.TempDir()  // テスト終了時に自動削除される
   ```

2. **実際のバイナリをビルドして実行する**
   ```go
   binaryPath := BuildBinary(t)
   output, err := ExecuteCommand(t, binaryPath, "command", "args")
   ```

3. **テーブル駆動テストを使用する**
   ```go
   tests := []struct {
       name    string
       args    []string
       wantErr bool
   }{
       {name: "正常系", args: []string{"init"}, wantErr: false},
       {name: "異常系", args: []string{"invalid"}, wantErr: true},
   }
   ```

4. **実行環境を汚さない**
   - 一時ディレクトリ内で作業する
   - テスト後にクリーンアップする（t.TempDirが自動的に行う）

5. **外部APIを使用するテストはスキップ可能にする**
   ```go
   geminiAPIKey := os.Getenv("GEMINI_API_KEY")
   if geminiAPIKey == "" {
       t.Skip("GEMINI_API_KEY環境変数が設定されていないためスキップします")
   }
   ```

### E2Eテストのトラブルシューティング

#### バイナリのビルドに失敗する

```bash
# プロジェクトルートで手動ビルドして確認
go build -o /tmp/ai-feed .
```

#### テストが一時ディレクトリ外で実行される

```go
// 明示的にディレクトリを変更し、クリーンアップを設定
originalWd, _ := os.Getwd()
os.Chdir(tmpDir)
t.Cleanup(func() {
    os.Chdir(originalWd)
})
```

#### E2Eテストが通常のテストで実行される

- ファイル先頭に`//go:build e2e`タグがあることを確認
- `go test`ではなく`go test -tags=e2e`で実行

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