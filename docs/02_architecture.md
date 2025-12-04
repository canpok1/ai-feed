# アーキテクチャ概要

ai-feedは、Cobraフレームワークを使用したGo CLIアプリケーションで、クリーンアーキテクチャパターンを採用しています。このドキュメントでは、プロジェクトの構造と設計の詳細を説明します。

## ディレクトリ構成

```
ai-feed/
├── cmd/                        # CLIコマンドの実装
│   ├── runner/                 # コマンドのビジネスロジック
│   │   └── *.go               # (profile_check, profile_init, recommend など)
│   ├── root.go                 # ルートコマンド
│   └── *.go                    # (init, profile, recommend, config など)
├── internal/                   # 内部パッケージ
│   ├── domain/                 # ドメイン層（ビジネスロジック）
│   │   ├── entity/             # エンティティ定義
│   │   │   └── *.go           # (config, entity, secret など)
│   │   ├── cache/              # キャッシュ実装
│   │   │   └── *.go           # (file_cache, nop_cache など)
│   │   ├── mock_domain/        # ドメイン層のモック（テスト用）
│   │   │   └── *.go           # 自動生成されるモック
│   │   └── *.go                # インターフェース定義 (comment, fetch, message など)
│   ├── infra/                  # インフラストラクチャ層
│   │   ├── comment/            # AI連携実装
│   │   │   └── *.go           # (factory, gemini など)
│   │   ├── fetch/              # フィード取得実装
│   │   │   └── *.go           # (rss など)
│   │   ├── message/            # メッセージ送信実装
│   │   │   └── *.go           # (builder, misskey, slack など)
│   │   ├── profile/            # プロファイル管理実装
│   │   │   └── *.go           # (repository など)
│   │   ├── selector/           # 記事選択実装
│   │   │   └── *.go           # (factory, gemini など)
│   │   ├── templates/          # 設定ファイルテンプレート
│   │   │   └── *.yml          # (config など)
│   │   ├── mock_infra/         # インフラ層のモック（テスト用）
│   │   │   └── *.go           # 自動生成されるモック
│   │   └── *.go                # (config, logger, templates など)
│   ├── testutil/               # テストユーティリティ
│   │   └── *.go               # 共通テストヘルパー
│   └── version/                # バージョン情報
│       └── *.go               # バージョン定義
├── test/                       # E2Eテスト
│   └── e2e/                    # E2Eテストコード
├── docs/                       # ドキュメント
├── main.go                     # エントリーポイント
├── go.mod                      # Goモジュール定義
├── go.sum                      # 依存関係のチェックサム
├── Makefile                    # ビルド・開発タスク
└── README.md                   # プロジェクト概要
```

## アーキテクチャパターン

### クリーンアーキテクチャの採用

本プロジェクトはクリーンアーキテクチャパターンを採用しています。これにより以下の利点を実現しています：

1. **ビジネスロジックの独立性**
   - ドメイン層（`internal/domain`）にビジネスロジックを集約
   - 外部システムへの依存を排除

2. **テスタビリティの向上**
   - インターフェースを介した疎結合な設計
   - モックを使用した単体テストが容易

3. **変更に対する柔軟性**
   - 外部サービスの変更が内部ロジックに影響しない
   - 新しい外部サービスの追加が容易

### レイヤー構成

```
┌──────────────────────────────────────────┐
│              Presentation Layer           │
│                   (cmd/)                  │
├──────────────────────────────────────────┤
│              Business Logic Layer         │
│              (cmd/runner/)                │
├──────────────────────────────────────────┤
│               Domain Layer                │
│           (internal/domain/)              │
├──────────────────────────────────────────┤
│          Infrastructure Layer             │
│            (internal/infra/)              │
└──────────────────────────────────────────┘
```

#### 1. Presentation Layer（cmd/）
- CLIコマンドの定義とパラメータ解析
- Cobraフレームワークを使用したコマンド構造の実装
- ユーザーインターフェースの提供

#### 2. Business Logic Layer（cmd/runner/）
- コマンドの実際のビジネスロジック実装
- ドメイン層とインフラ層の協調
- エラーハンドリングとユーザーフィードバック

#### 3. Domain Layer（internal/domain/）
- コアビジネスロジックとルール
- エンティティとバリューオブジェクトの定義
- 外部依存のないピュアな実装

#### 4. Infrastructure Layer（internal/infra/）
- 外部サービスとの連携実装
- ファイルシステムやネットワークアクセス
- ドメイン層のインターフェース実装

## 依存性注入パターン

### 概要
依存性注入（DI）パターンを採用し、テスタブルで保守性の高いコードを実現しています。

### 実装例

```go
// インターフェース定義（ドメイン層）
type Fetcher interface {
    Fetch(ctx context.Context, urls []string) ([]entity.Article, error)
}

// 実装（インフラ層）
type RSSFetcher struct {
    client *http.Client
}

// コンストラクタインジェクション
func NewRecommendRunner(
    fetcher domain.Fetcher,
    recommender domain.Recommender,
    viewers []domain.Viewer,
) *RecommendRunner {
    return &RecommendRunner{
        fetcher:     fetcher,
        recommender: recommender,
        viewers:     viewers,
    }
}
```

### 利点
- モックを使用した単体テストが容易
- 実装の切り替えが簡単
- 依存関係が明確

## 主要なアーキテクチャ上の決定事項

### 1. 設定システム

**決定**: YAMLファイルベースの階層的設定システム

**理由**:
- 人間が読み書きしやすい形式
- 階層的な設定の表現が自然
- プロファイルによる設定の切り替えが容易

**実装**:
- `config.yml`: グローバル設定とデフォルトプロファイル
- `profile.yml`: 個別のプロファイル設定
- マージ機能による設定の上書き

### 2. AI連携アーキテクチャ

**決定**: ファクトリーパターンによるAIプロバイダーの抽象化

**理由**:
- 複数のAIプロバイダーへの対応が容易
- プロバイダー固有の実装を隠蔽
- 将来的な拡張性の確保

**現在の実装**:
- Google Gemini APIのサポート
- インターフェースによる抽象化
- プロンプトテンプレートのカスタマイズ

### 3. テスト戦略

**決定**: テーブル駆動テストとモック生成の組み合わせ

**理由**:
- テストケースの管理が容易
- 網羅的なテストの実現
- 実装とテストの分離

**ツール**:
- testifyフレームワーク: アサーションとテストスイート
- go.uber.org/mock: モック生成
- テーブル駆動テスト: 複数のテストケースを効率的に管理

### 4. 外部連携の設計

**決定**: Viewerパターンによる出力先の抽象化

**理由**:
- 複数の出力先への同時送信
- 新しい出力先の追加が容易
- エラーハンドリングの統一

**対応プラットフォーム**:
- Slack（Webhook API）
- Misskey（REST API）
- 標準出力（ログ）

### 5. ログシステム

**決定**: Go標準ライブラリのslogパッケージを採用

**理由**:
- 構造化ログの標準的なサポート
- パフォーマンスの良さ
- 外部依存の削減

**機能**:
- レベルベースのログ出力（DEBUG/INFO/WARN/ERROR）
- 色付き出力によるk視認性向上
- verboseフラグによる詳細度の切り替え

## コマンド設計パターン

### Runner パターン

各コマンドのビジネスロジックは`cmd/runner`パッケージに分離されています：

```go
// コマンド定義（cmd/recommend.go）
func makeRecommendCmd(...) *cobra.Command {
    return &cobra.Command{
        RunE: func(cmd *cobra.Command, args []string) error {
            // パラメータ解析
            params := newRecommendParams(cmd)
            
            // Runner実行
            runner := runner.NewRecommendRunner(...)
            return runner.Run(cmd.Context(), params, profile)
        },
    }
}

// ビジネスロジック（cmd/runner/recommend.go）
type RecommendRunner struct {
    fetcher     domain.Fetcher
    recommender domain.Recommender
    viewers     []domain.Viewer
}

func (r *RecommendRunner) Run(ctx context.Context, params *RecommendParams, profile Profile) error {
    // 実際の処理
}
```

### パラメータ構造体

コマンドパラメータは専用の構造体で管理：

```go
type RecommendParams struct {
    URLs   []string
    Source string
}
```

## エラーハンドリング

### Sentinel Error パターン

特定の状況を表すエラーをパッケージレベルで定義：

```go
// エラー定義
var ErrNoArticlesFound = errors.New("no articles found in the feed")

// エラー判定
if errors.Is(err, runner.ErrNoArticlesFound) {
    // ユーザーフレンドリーなメッセージを表示
    return nil
}
```

### エラーの集約

複数の処理でのエラーを集約：

```go
var errs []error
for _, viewer := range viewers {
    if err := viewer.View(article); err != nil {
        errs = append(errs, fmt.Errorf("viewer error: %w", err))
    }
}
return errors.Join(errs...)
```

## 拡張ポイント

### 新しいAIプロバイダーの追加

1. `internal/domain/comment.go`のインターフェースを実装
2. `internal/infra/comment/`に新しい実装を追加
3. ファクトリーに登録

### 新しい出力先の追加

1. `internal/domain/message.go`のインターフェースを実装
2. `internal/infra/message/`に新しい実装を追加
3. 設定構造体を拡張

### 新しいコマンドの追加

1. `cmd/`に新しいコマンドファイルを作成
2. `cmd/runner/`にビジネスロジックを実装
3. `cmd/root.go`でコマンドを登録

## パフォーマンス考慮事項

### 並行処理

複数のフィードやビューアーの処理で適切に並行処理を活用：

```go
// 複数のビューアーへの並行送信
var wg sync.WaitGroup
for _, viewer := range viewers {
    wg.Add(1)
    go func(v Viewer) {
        defer wg.Done()
        v.SendRecommend(article)
    }(viewer)
}
wg.Wait()
```

### リソース管理

- HTTPクライアントの再利用
- 適切なタイムアウト設定
- メモリ効率的なストリーム処理

## セキュリティ考慮事項

### 認証情報の管理

- APIキーは環境変数または設定ファイルで管理
- 設定ファイルは`.gitignore`で除外
- トークンのローテーションを推奨

### 入力検証

- URLの妥当性チェック
- プロンプトインジェクション対策
- サニタイゼーション処理

## まとめ

ai-feedのアーキテクチャは、クリーンアーキテクチャの原則に従い、保守性、テスタビリティ、拡張性を重視して設計されています。各層の責任が明確に分離されており、新しい機能の追加や既存機能の変更が容易な構造となっています。