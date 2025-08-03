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

## 10. コミット前チェックリスト

1. `make fmt` - コードフォーマット
2. `make lint` - 静的解析
3. `make test` - テスト実行
4. `go mod tidy` - 依存関係の整理（新規追加時）

これらのルールは、コードの可読性、保守性、拡張性を高めることを目的としています。新しいコードを書く際は、既存のコードベースのスタイルに合わせることを心がけてください。
