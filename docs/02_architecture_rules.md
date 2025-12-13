# アーキテクチャ概要

ai-feedは、Cobraフレームワークを使用したGo CLIアプリケーションで、クリーンアーキテクチャパターンを採用しています。このドキュメントでは、プロジェクトの構造と設計の詳細を説明します。

## ディレクトリ構成

```
ai-feed/
├── cmd/                        # Presentation Layer: CLIコマンドの定義
│   ├── root.go                 # ルートコマンド
│   └── *.go                    # (init, profile, recommend, config など)
├── internal/                   # 内部パッケージ
│   ├── app/                    # Application Layer: ユースケース
│   │   ├── mock_app/           # アプリケーション層のモック（テスト用）
│   │   │   └── *.go           # 自動生成されるモック
│   │   └── *.go               # (recommend, profile_check, config_check など)
│   ├── domain/                 # Domain Layer: ビジネスルール
│   │   ├── entity/             # エンティティ定義
│   │   │   └── *.go           # (config, entity, secret など)
│   │   ├── mock_domain/        # ドメイン層のモック（テスト用）
│   │   │   └── *.go           # 自動生成されるモック
│   │   └── *.go                # インターフェース定義 (comment, fetch, message など)
│   ├── infra/                  # Infrastructure Layer: 外部連携
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
│   │   ├── cache/              # キャッシュ実装
│   │   │   └── *.go           # (file_cache, nop_cache など)
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
   - ドメイン層（`internal/domain`）にビジネスルールを集約
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
│          Presentation Layer              │
│                (cmd/)                    │
├──────────────────────────────────────────┤
│          Application Layer               │
│            (internal/app/)               │
├──────────────────────────────────────────┤
│            Domain Layer                  │
│          (internal/domain/)              │
├──────────────────────────────────────────┤
│        Infrastructure Layer              │
│          (internal/infra/)               │
└──────────────────────────────────────────┘
```

### 依存関係ルール

```
cmd → app → domain ← infra
              ↑
         依存性逆転
```

- **上位層は下位層に依存できる**（cmd → app → domain）
- **infra層はdomain層のインターフェースを実装する**（依存性逆転）
- **domain層は他のどの層にも依存しない**（最も内側の層）

## 各層の責務定義

### 1. Presentation Layer（cmd/）

**役割**: ユーザーインターフェースの提供

**責務**:
- CLIコマンドの定義（Cobraフレームワーク）
- コマンドライン引数・フラグの解析
- Application層への処理委譲
- ユーザー向けエラーメッセージの表示（Application層からのエラーを変換）

**許可される依存**:
- `internal/app`（Application層）
- `github.com/spf13/cobra`（CLIフレームワーク）

**禁止事項**:
- ❌ domain層への直接依存（entity参照を除く）
- ❌ infra層への直接依存
- ❌ ビジネスロジックの実装
- ❌ 設定ファイルの読み込み・パース処理
- ❌ 外部サービスとの直接通信

**コード例**:
```go
// cmd/recommend.go - 良い例
func makeRecommendCmd() *cobra.Command {
    return &cobra.Command{
        Use:   "recommend",
        Short: "記事を推薦します",
        RunE: func(cmd *cobra.Command, args []string) error {
            // フラグ解析のみ
            params := parseRecommendFlags(cmd)

            // Application層に委譲
            useCase := app.NewRecommendUseCase(/* 依存性注入 */)
            return useCase.Execute(cmd.Context(), params)
        },
    }
}
```

### 2. Application Layer（internal/app/）

**役割**: ユースケースの実装とオーケストレーション

**責務**:
- ユースケースの実装（各コマンドの処理フロー）
- 設定ファイルの読み込み・マージ・検証の調整
- 依存オブジェクトの組み立て（DIワイヤリング）
- domain層とinfra層の協調
- トランザクション的な処理の管理
- ユーザーへの進行状況フィードバック

**許可される依存**:
- `internal/domain`（Domain層）
- `internal/infra`（Infrastructure層）※DIのための参照のみ

**禁止事項**:
- ❌ CLIフレームワーク（Cobra）への依存
- ❌ infra層の具象型を直接フィールドに保持（domainインターフェース経由で保持）
- ❌ 外部APIの直接呼び出し

**コード例**:
```go
// internal/app/recommend.go - 良い例
type RecommendUseCase struct {
    fetcher     domain.FetchClient      // domainインターフェース
    recommender domain.Recommender      // domainインターフェース
    senders     []domain.MessageSender  // domainインターフェース
}

func (u *RecommendUseCase) Execute(ctx context.Context, params RecommendParams) error {
    // 1. 設定の読み込み・マージ
    // 2. バリデーション
    // 3. ビジネスロジックの実行（domain層経由）
    // 4. 結果の通知（domain層インターフェース経由）
}
```

### 3. Domain Layer（internal/domain/）

**役割**: ビジネスルールとコアロジックの定義

**責務**:
- インターフェースの定義（Fetcher, Recommender, MessageSender等）
- エンティティ・値オブジェクトの定義（entity/配下）
- ドメインサービスの実装（純粋なビジネスロジックのみ）
- バリデーションルール

**許可される依存**:
- 標準ライブラリのみ
- `internal/domain/entity`（同一層内）

**禁止事項**:
- ❌ 他の層（cmd, app, infra）への依存
- ❌ 外部ライブラリへの依存（標準ライブラリ以外）
- ❌ I/O処理（ファイル、ネットワーク、DB）
- ❌ ログ出力（slog等）を含む副作用

**コード例**:
```go
// internal/domain/recommend.go - 良い例
type Recommender interface {
    Recommend(ctx context.Context, articles []entity.Article) (*entity.Recommend, error)
}

type ArticleSelector interface {
    Select(ctx context.Context, articles []entity.Article) (*entity.Article, error)
}

// internal/domain/entity/article.go - 良い例
type Article struct {
    Title     string
    Link      string
    Published *time.Time
}

func (a *Article) Validate() error {
    if a.Link == "" {
        return errors.New("link is required")
    }
    return nil
}
```

### 4. Infrastructure Layer（internal/infra/）

**役割**: 外部システムとの連携実装

**責務**:
- domain層インターフェースの具象実装
- 外部API呼び出し（Gemini, Slack, Misskey等）
- ファイルI/O（設定ファイル、キャッシュ）
- ネットワーク通信（RSS取得等）
- ログ出力

**許可される依存**:
- `internal/domain`（Domain層）
- 外部ライブラリ（slack-go, genai等）

**禁止事項**:
- ❌ cmd層への依存
- ❌ app層への依存
- ❌ ビジネスルールの実装（domain層の責務）

**コード例**:
```go
// internal/infra/fetch/rss.go - 良い例
type RSSFetcher struct {
    client *http.Client
}

// domain.FetchClient インターフェースを実装
func (f *RSSFetcher) Fetch(url string) ([]entity.Article, error) {
    // RSS取得の実装
}

// internal/infra/message/slack.go - 良い例
type SlackSender struct {
    client  *slack.Client
    channel string
}

// domain.MessageSender インターフェースを実装
func (s *SlackSender) SendRecommend(r *entity.Recommend, msg string) error {
    // Slack送信の実装
}
```

## 層間の依存関係まとめ

| 層 | 依存できる層 | 依存できない層 |
|---|---|---|
| cmd (Presentation) | app | domain（entity除く）, infra |
| app (Application) | domain, infra（DI用） | cmd |
| domain (Domain) | なし（標準ライブラリのみ） | cmd, app, infra |
| infra (Infrastructure) | domain | cmd, app |

## 依存性注入パターン

### 概要
依存性注入（DI）パターンを採用し、テスタブルで保守性の高いコードを実現しています。

### 実装例

```go
// インターフェース定義（Domain層）
type FetchClient interface {
    Fetch(url string) ([]entity.Article, error)
}

type Recommender interface {
    Recommend(ctx context.Context, articles []entity.Article) (*entity.Recommend, error)
}

// 実装（Infrastructure層）
type RSSFetcher struct {
    client *http.Client
}

func (f *RSSFetcher) Fetch(url string) ([]entity.Article, error) {
    // 実装
}

// コンストラクタインジェクション（Application層）
func NewRecommendUseCase(
    fetcher domain.FetchClient,
    recommender domain.Recommender,
    senders []domain.MessageSender,
) *RecommendUseCase {
    return &RecommendUseCase{
        fetcher:     fetcher,
        recommender: recommender,
        senders:     senders,
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

### UseCase パターン

各コマンドのビジネスロジックは`internal/app`パッケージにUseCaseとして分離されています：

```go
// コマンド定義（cmd/recommend.go）
func makeRecommendCmd() *cobra.Command {
    return &cobra.Command{
        RunE: func(cmd *cobra.Command, args []string) error {
            // パラメータ解析のみ
            params := parseRecommendFlags(cmd)

            // UseCase実行（DIはUseCase内部で行う）
            useCase := app.NewRecommendUseCase()
            return useCase.Execute(cmd.Context(), params)
        },
    }
}

// ユースケース（internal/app/recommend.go）
type RecommendUseCase struct {
    fetcher     domain.FetchClient
    recommender domain.Recommender
    senders     []domain.MessageSender
}

func (u *RecommendUseCase) Execute(ctx context.Context, params *RecommendParams) error {
    // 1. 設定読み込み・マージ
    // 2. バリデーション
    // 3. ビジネスロジック実行
    // 4. 結果通知
}
```

### パラメータ構造体

コマンドパラメータは専用の構造体で管理：

```go
// internal/app/recommend.go
type RecommendParams struct {
    URLs        []string
    ConfigPath  string
    ProfilePath string
}
```

## エラーハンドリング

### Sentinel Error パターン

特定の状況を表すエラーをパッケージレベルで定義：

```go
// エラー定義（internal/app/recommend.go）
var ErrNoArticlesFound = errors.New("no articles found in the feed")

// エラー判定（cmd/recommend.go）
if errors.Is(err, app.ErrNoArticlesFound) {
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

1. `cmd/`に新しいコマンドファイルを作成（フラグ解析のみ）
2. `internal/app/`にユースケースを実装
3. `cmd/root.go`でコマンドを登録

## パフォーマンス考慮事項

### 並行処理

複数のフィードやメッセージ送信の処理で適切に並行処理を活用：

```go
// 複数のMessageSenderへの並行送信
var wg sync.WaitGroup
for _, sender := range senders {
    wg.Add(1)
    go func(s domain.MessageSender) {
        defer wg.Done()
        s.SendRecommend(recommend, fixedMessage)
    }(sender)
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