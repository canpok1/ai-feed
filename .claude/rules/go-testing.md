---
paths: "**/*_test.go"
---
# Go テストルール

このルールは `*_test.go` ファイルに適用されます。
詳細は [docs/03_testing_rules.md](../../docs/03_testing_rules.md) を参照してください。

## テストフレームワーク

| ツール | 用途 |
|--------|------|
| testify | アサーション（`assert`/`require`） |
| go.uber.org/mock | モック生成・管理 |

## 層別カバレッジ目標

| 層 | 目標 | テスト種類 |
|---|---|---|
| domain | 80%以上 | ユニット + 統合 |
| infra | 60%以上 | ユニット + 統合 |
| app | 50%以上 | ユニット + 統合 |
| cmd | E2Eで担保 | E2Eテスト |

## テーブル駆動テストパターン

```go
func TestXxx(t *testing.T) {
    tests := []struct {
        name    string
        input   InputType
        want    OutputType
        wantErr bool
    }{
        {name: "正常系: ...", input: ..., want: ..., wantErr: false},
        {name: "異常系: ...", input: ..., want: ..., wantErr: true},
    }

    for _, tt := range tests {
        t.Run(tt.name, func(t *testing.T) {
            // テスト実装
        })
    }
}
```

## アサーションの使い分け

| 関数 | 用途 |
|------|------|
| `assert.Equal` | 値の等価性確認（テスト継続） |
| `require.NoError` | エラーなし確認（失敗時即座に終了） |
| `assert.ErrorIs` | Sentinel error の確認 |

## ベストプラクティス

- 一時ディレクトリは `t.TempDir()` を使用
- モック再生成: `make generate`
- 統合テストには `//go:build integration` タグを付与
- テスト名は内容を明確に表す（例: `正常系: 記事取得成功`）
