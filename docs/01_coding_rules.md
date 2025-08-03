# ai-feed コーディングルール

このドキュメントは、ai-feedプロジェクトにおけるGoコードの記述規約をまとめたものです。

## 1. 命名規則

### 1.1 基本ルール
- **パッケージ名**: 小文字のみ使用 (例: `cmd`, `domain`, `infra`)
- **パブリック関数/型**: パスカルケース (例: `NewFetcher`, `Article`)
- **プライベート関数/変数**: キャメルケース (例: `makeRootCmd`, `fetchClient`)
- **定数**: パスカルケース (例: `DefaultTimeout`)

### 1.2 構造体とインターフェース
```go
// Good: インターフェースは動詞または-erで終わる
type Fetcher interface {
    Fetch(url string) ([]entity.Article, error)
}

// Good: 構造体は名詞
type Article struct {
    Title       string `json:"title"`
    Link        string `json:"link"`
    Description string `json:"description"`
}
```

### 1.3 ファイル名
- 機能を表す名前を使用: `config.go`, `fetch.go`
- テストファイル: `{対象ファイル名}_test.go`
- モックファイル: `mock_{パッケージ名}/` ディレクトリに配置

## 2. エラーハンドリング

### 2.1 エラーのラップ
```go
// Good: コンテキストを含むエラーメッセージ
if err != nil {
    return fmt.Errorf("failed to load config from %s: %w", configPath, err)
}
```

### 2.2 Nilチェック
```go
// Good: 明確なエラーメッセージ
if writer == nil {
    return nil, fmt.Errorf("writer cannot be nil")
}
```

### 2.3 エラー集約
```go
// Good: 複数のエラーを集約
var errs []error
for _, viewer := range viewers {
    if err := viewer.View(article); err != nil {
        errs = append(errs, fmt.Errorf("viewer error: %w", err))
    }
}
return errors.Join(errs...)
```

## 3. コメント規約

### 3.1 言語の使い分け
- **日本語**: 全てのコメント

## 4. パッケージ構造

### 4.1 レイヤード・アーキテクチャ
```
cmd/                # CLI コマンド実装
internal/
  domain/           # ビジネスロジック
    entity/         # ドメインエンティティ
  infra/            # 外部サービス連携
```

### 4.2 インターフェースの配置
- ドメイン層でインターフェースを定義
- インフラ層で実装を提供

## 5. 依存性注入

### 5.1 コンストラクタパターン
```go
// Good: 依存関係を明示的に注入
func NewRecommender(generator CommentGenerator) *Recommender {
    return &Recommender{
        generator: generator,
    }
}
```

### 5.2 ファクトリーパターン
```go
// Good: 複雑な初期化にはファクトリーを使用
type CommentGeneratorFactory interface {
    MakeCommentGenerator(config *AIConfig) (CommentGenerator, error)
}
```

## 6. テストコード

### 6.1 テーブル駆動テスト
```go
func TestFetch(t *testing.T) {
    tests := []struct {
        name        string
        setupMock   func(*mock_domain.MockClient)
        want        []entity.Article
        wantErr     bool
    }{
        {
            name: "正常系: 記事取得成功",
            setupMock: func(m *mock_domain.MockClient) {
                m.EXPECT().Get(gomock.Any()).Return(testArticles, nil)
            },
            want:    testArticles,
            wantErr: false,
        },
    }
    
    for _, tt := range tests {
        t.Run(tt.name, func(t *testing.T) {
            // テスト実装
        })
    }
}
```

### 6.2 モックの使用
- `go.uber.org/mock` を使用
- `make generate` でモックを自動生成
- モックはインターフェース毎に作成

### 6.3 一時ファイル
```go
// Good: t.TempDir()を使用
tmpDir := t.TempDir()
configPath := filepath.Join(tmpDir, "test.yaml")
```

## 7. 設定管理

### 7.1 YAML構造体タグ
```go
type Config struct {
    AI     *AIConfig     `yaml:"ai"`
    Output *OutputConfig `yaml:"output"`
}
```

### 7.2 マージパターン
```go
// Good: nilチェックを含むマージ関数
func (p *Profile) Merge(other *Profile) {
    if other == nil {
        return
    }
    mergePtr(&p.AI, other.AI)
}
```

## 8. ジェネリクス

### 8.1 型安全な汎用関数
```go
// Good: 再利用可能な汎用関数
func loadYaml[T any](filePath string) (*T, error) {
    // 実装
}
```

## 9. その他のベストプラクティス

### 9.1 早期リターン
```go
// Good: ネストを減らす
if err != nil {
    return err
}
// 正常系の処理
```

### 9.2 小さなインターフェース
```go
// Good: 単一責任の原則
type Viewer interface {
    View(article entity.Article) error
}
```

### 9.3 定数の使用
- マジックナンバーは避ける
- 設定可能な値はconfigに移動

## 10. cmd/runner パッケージの設計パターン

### 10.1 ビジネスロジック分離の原則
CLIコマンドのビジネスロジックは`cmd/runner`パッケージに分離し、`cmd`パッケージはCLIの引数解析と依存性注入のみを担当する。

```go
// Good: cmd/runner/recommend.go
type RecommendRunner struct {
    fetcher     *domain.Fetcher
    recommender domain.Recommender
    viewers     []domain.Viewer
}

func (r *RecommendRunner) Run(ctx context.Context, params *RecommendParams, profile infra.Profile) error {
    // ビジネスロジックの実装
}

// Good: cmd/recommend.go
func makeRecommendCmd(fetchClient domain.FetchClient, recommender domain.Recommender) *cobra.Command {
    cmd := &cobra.Command{
        RunE: func(cmd *cobra.Command, args []string) error {
            // 引数解析と依存性注入のみ
            recommendRunner, err := runner.NewRecommendRunner(...)
            return recommendRunner.Run(cmd.Context(), params, currentProfile)
        },
    }
}
```

### 10.2 パラメータ構造体パターン
コマンドの実行パラメータは専用の構造体で定義し、コマンドライン引数と分離する。

```go
// Good: 実行パラメータの構造体
type RecommendParams struct {
    URLs []string
}

// Good: パラメータ作成関数
func newRecommendParams(cmd *cobra.Command) (*runner.RecommendParams, error) {
    // フラグからパラメータを構築
}
```

## 11. Sentinel Error パターン

### 11.1 Sentinel Error の定義
特定の状況を表すエラーはsentinel errorとして定義し、公開パッケージレベルで管理する。

```go
// Good: パッケージレベルでsentinel errorを定義
var ErrNoArticlesFound = errors.New("no articles found in the feed")
```

### 11.2 Sentinel Error の使用
sentinel errorの判定には`errors.Is`を使用し、型アサーションは避ける。

```go
// Good: errors.Isを使用した判定
if errors.Is(err, runner.ErrNoArticlesFound) {
    fmt.Fprintln(cmd.OutOrStdout(), "記事が見つかりませんでした。")
    return nil
}

// Bad: 文字列比較
if err.Error() == "no articles found in the feed" {
    // エラーメッセージが変更されると動作しなくなる
}
```

### 11.3 コメント規約
sentinel errorには必ずその用途を説明する日本語コメントを付ける。

```go
// ErrNoArticlesFound は記事が見つからなかった場合のsentinel error
var ErrNoArticlesFound = errors.New("no articles found in the feed")
```

## 12. 変数名の競合回避パターン

### 12.1 err変数の再利用回避
同一スコープ内で`err`変数を再利用することは避け、明確な変数名を使用する。

```go
// Good: 明確な変数名を使用
err := r.recommender.Recommend(ctx, aiConfig, promptConfig, allArticles)
if err != nil {
    return fmt.Errorf("failed to recommend article: %w", err)
}

var errs []error
for _, viewer := range r.viewers {
    if viewErr := viewer.ViewRecommend(recommend, profile.Prompt.FixedMessage); viewErr != nil {
        errs = append(errs, fmt.Errorf("failed to view recommend: %w", viewErr))
    }
}

// Bad: err変数の再利用
err := r.recommender.Recommend(ctx, aiConfig, promptConfig, allArticles)
if err != nil {
    return fmt.Errorf("failed to recommend article: %w", err)
}

for _, viewer := range r.viewers {
    if err := viewer.ViewRecommend(recommend, profile.Prompt.FixedMessage); err != nil {
        // errが再利用されており、可読性が低い
    }
}
```

### 12.2 パッケージ名と変数名の競合回避
変数名がパッケージ名と競合する場合は、より具体的な変数名を使用する。

```go
// Good: 具体的な変数名を使用
recommendRunner, runnerErr := runner.NewRecommendRunner(...)

// Bad: パッケージ名と同じ変数名
runner, runnerErr := runner.NewRecommendRunner(...)
```

## 13. Context の使用パターン

### 13.1 Context の引数位置
`context.Context`は常に関数の最初の引数として配置する。

```go
// Good: contextが最初の引数
func (r *RecommendRunner) Run(ctx context.Context, params *RecommendParams, profile infra.Profile) error

// Good: メソッドでもcontextが最初
func (g *geminiCommentGenerator) Generate(ctx context.Context, article *entity.Article) (string, error)

// Bad: contextが最初以外の位置
func (r *RecommendRunner) Run(params *RecommendParams, ctx context.Context, profile infra.Profile) error
```

### 13.2 Context の伝播
上位層から下位層にcontextを適切に伝播させる。

```go
// Good: コマンドレベルからビジネスロジックにcontextを伝播
func makeRecommendCmd(...) *cobra.Command {
    cmd := &cobra.Command{
        RunE: func(cmd *cobra.Command, args []string) error {
            return recommendRunner.Run(cmd.Context(), params, currentProfile)
        },
    }
}
```

## 14. 高度なエラーハンドリング

### 14.1 複数エラーの集約
複数の処理でエラーが発生する可能性がある場合は、エラーを集約して処理を継続する。

```go
// Good: エラー集約パターン
var errs []error
for _, viewer := range r.viewers {
    if viewErr := viewer.ViewRecommend(recommend, profile.Prompt.FixedMessage); viewErr != nil {
        errs = append(errs, fmt.Errorf("failed to view recommend: %w", viewErr))
    }
}

if len(errs) > 0 {
    return fmt.Errorf("failed to view all recommends: %v", errs)
}
```

### 14.2 ユーザーフレンドリーなエラー処理
特定のエラーに対してはユーザーフレンドリーなメッセージを表示し、正常終了扱いにする。

```go
// Good: ユーザーフレンドリーなエラー処理
err = recommendRunner.Run(cmd.Context(), params, currentProfile)
if err != nil {
    if errors.Is(err, runner.ErrNoArticlesFound) {
        fmt.Fprintln(cmd.OutOrStdout(), "記事が見つかりませんでした。")
        return nil  // エラーではなく正常終了
    }
    return err
}
```

## 15. マージ処理パターン

### 15.1 階層的マージ
設定やプロファイルのマージでは、nil チェックと階層的なマージを実装する。

```go
// Good: 階層的マージパターン
func (p *Profile) Merge(other *Profile) {
    if other == nil {
        return
    }
    mergePtr(&p.AI, other.AI)
    mergePtr(&p.Prompt, other.Prompt)  
    mergePtr(&p.Output, other.Output)
}
```

### 15.2 デフォルト値との組み合わせ
設定ロジックでは、デフォルト値、設定ファイル、コマンドライン引数の優先順位を明確にする。

```go
// Good: 優先順位の明確な設定ロジック
var currentProfile infra.Profile

// 1. デフォルト値を設定
if config.DefaultProfile != nil {
    currentProfile = *config.DefaultProfile
}

// 2. プロファイルファイルでマージ（オーバーライド）
if profilePath != "" {
    loadedProfile, err := infra.NewYamlProfileRepository(profilePath).LoadProfile()
    if err != nil {
        return fmt.Errorf("failed to load profile from %s: %w", profilePath, err)
    }
    currentProfile.Merge(loadedProfile)
}
```

## 16. コミット前チェックリスト

1. `make fmt` - コードフォーマット
2. `make lint` - 静的解析
3. `make test` - テスト実行
4. `go mod tidy` - 依存関係の整理（新規追加時）

これらのルールは、コードの可読性、保守性、拡張性を高めることを目的としています。新しいコードを書く際は、既存のコードベースのスタイルに合わせることを心がけてください。
