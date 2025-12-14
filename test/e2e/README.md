# E2Eテストガイド

このディレクトリには、ai-feedのE2E（End-to-End）テストが配置されています。

## 概要

E2Eテストは、実際のバイナリをビルドして実行し、エンドユーザーの視点で動作を検証するテストです。

## ディレクトリ構造

```
test/e2e/
├── common/                        # 共通ユーティリティ
│   ├── helper.go                  # 共通ヘルパー関数
│   └── mock/                      # モックサーバー実装
│       ├── rss.go                 # RSS/Atomフィードモックサーバー
│       ├── slack.go               # Slackモックサーバー
│       └── misskey.go             # Misskeyモックサーバー
├── config/                        # configコマンドのテスト
│   ├── config_test.go
│   ├── main_test.go
│   └── testdata/
├── init/                          # initコマンドのテスト
│   ├── init_test.go
│   └── main_test.go
├── profile/                       # profileコマンドのテスト
│   ├── profile_test.go
│   ├── main_test.go
│   └── testdata/
└── recommend/                     # recommendコマンドのテスト
    ├── recommend_test.go
    ├── main_test.go
    └── testdata/
```

## 実行方法

```bash
# 全E2Eテストを実行
make test-e2e

# 直接実行
go test -tags=e2e -v ./test/e2e/...

# 特定のコマンドのテストのみ実行
go test -tags=e2e -v ./test/e2e/init/
```

## ビルドタグ

E2Eテストファイルには `//go:build e2e` タグを追加します。これにより、通常の `go test` では実行されません。

```go
//go:build e2e

package e2e
```

## ヘルパー関数

`test/e2e/common/helper.go` に以下のヘルパー関数が定義されています：

| 関数 | 説明 |
|------|------|
| `GetProjectRoot(t)` | プロジェクトルートディレクトリのパスを取得 |
| `BuildBinary(t)` | テスト用バイナリをビルドし、パスを返す |
| `ExecuteCommand(t, binaryPath, args...)` | バイナリを実行し、出力を返す |
| `CreateTestConfig(t, tmpDir, params)` | テスト用設定ファイルを作成 |

## main_test.go の役割

各コマンドディレクトリの `main_test.go` は `TestMain` 関数を定義し、以下を担当します：

- テスト実行前のセットアップ（バイナリビルド等）
- `sync.Once` による複数パッケージ実行時の一度だけのビルド
- テスト終了後のクリーンアップ

```go
func TestMain(m *testing.M) {
    common.SetupPackage()
    code := m.Run()
    common.CleanupPackage()
    os.Exit(code)
}
```

## AIモック機能

外部AI API（Gemini）への依存を排除するためのモック機能を提供しています。

### 設定オプション

| 設定項目 | 説明 | デフォルト値 |
|----------|------|--------------|
| `ai.mock.enabled` | モック機能の有効/無効 | `false` |
| `ai.mock.selector_mode` | 記事選択モード（`first`, `random`, `last`） | `first` |
| `ai.mock.comment` | モックが返す固定コメント | 空文字列 |

### 設定例

```yaml
default_profile:
  ai:
    mock:
      enabled: true
      selector_mode: first
      comment: "これはテスト用のモックコメントです。"
```

## モックサーバー

### RSSフィードモック (`common/mock/rss.go`)

| ハンドラ | 説明 |
|----------|------|
| `NewMockRSSHandler()` | 標準的なRSSフィードを返す |
| `NewMockAtomHandler()` | Atomフィードを返す |
| `NewMockEmptyFeedHandler()` | 空のフィードを返す |
| `NewMockInvalidFeedHandler()` | 不正なXMLを返す |
| `NewMockErrorHandler(statusCode)` | 指定したHTTPエラーを返す |

### Slackモック (`common/mock/slack.go`)

| メソッド | 説明 |
|----------|------|
| `ReceivedMessage()` | メッセージ受信確認 |
| `GetMessages()` | 受信したメッセージ一覧取得 |
| `Reset()` | 状態リセット |

### Misskeyモック (`common/mock/misskey.go`)

| メソッド | 説明 |
|----------|------|
| `ReceivedNote()` | ノート受信確認 |
| `GetNotes()` | 受信したノート一覧取得 |
| `Reset()` | 状態リセット |

## テストの書き方

```go
//go:build e2e

package init

import (
    "os"
    "testing"

    "github.com/stretchr/testify/assert"
    "github.com/canpok1/ai-feed/test/e2e/common"
)

func TestInitCommand_CreateConfigFile(t *testing.T) {
    // バイナリをビルド
    binaryPath := common.BuildBinary(t)

    // 一時ディレクトリを作成
    tmpDir := t.TempDir()
    originalWd, _ := os.Getwd()
    os.Chdir(tmpDir)
    t.Cleanup(func() { os.Chdir(originalWd) })

    // コマンドを実行
    output, err := common.ExecuteCommand(t, binaryPath, "init")

    // 結果を検証
    assert.NoError(t, err)
    assert.Contains(t, output, "config.yml を生成しました")

    // ファイルが作成されたことを確認
    _, err = os.Stat("config.yml")
    assert.NoError(t, err)
}
```

## recommendコマンドのテスト

recommendコマンドは外部API（Gemini）を使用するため、以下の2つの方法でテストできます：

### 1. AIモックを使用（推奨）

```go
config := common.CreateRecommendTestConfig(t, tmpDir, common.RecommendConfigParams{
    FeedURLs:        []string{feedURL},
    SlackWebhookURL: slackURL,
    // UseMockAI: デフォルトでtrue
})
```

### 2. 実際のGemini APIを使用

```go
useMockAI := false
config := common.CreateRecommendTestConfig(t, tmpDir, common.RecommendConfigParams{
    UseMockAI:       &useMockAI,
    GeminiAPIKey:    os.Getenv("GEMINI_API_KEY"),
    FeedURLs:        []string{feedURL},
    SlackWebhookURL: slackURL,
})
```

## GitHub ActionsでのE2Eテスト

`GEMINI_API_KEY` を Repository Secrets に設定することで、CIでも実際のAPIを使用したテストが可能です。

1. GitHub リポジトリ > Settings > Secrets and variables > Actions
2. "New repository secret" をクリック
3. Name: `GEMINI_API_KEY`、Secret: Gemini APIキーを入力
4. "Add secret" をクリック

設定後、プルリクエストやmainブランチへのプッシュ時に自動的にE2Eテストが実行されます。

## ベストプラクティス

1. **一時ディレクトリを使用**: `t.TempDir()` を使用（テスト終了時に自動削除）
2. **テーブル駆動テストを使用**: 複数のテストケースを効率的に管理
3. **実行環境を汚さない**: 一時ディレクトリ内で作業
4. **外部APIテストはスキップ可能に**: 環境変数未設定時に `t.Skip()` を使用

## トラブルシューティング

### バイナリのビルド失敗

```bash
# プロジェクトルートで手動ビルドして確認
go build -o /tmp/ai-feed .
```

### E2Eテストが通常のテストで実行される

- ファイル先頭に `//go:build e2e` タグがあることを確認
- `go test -tags=e2e` で実行
