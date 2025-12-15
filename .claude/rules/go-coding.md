---
paths: "**/*.go"
---
# Go コーディング・アーキテクチャルール

このルールは Go ファイル（テストファイルを除く）に適用されます。
詳細は [docs/01_coding_rules.md](../../docs/01_coding_rules.md) および [docs/02_architecture_rules.md](../../docs/02_architecture_rules.md) を参照してください。

## 命名規則

- パッケージ名: 小文字のみ（例: `cmd`, `domain`, `infra`）
- パブリック関数/型: パスカルケース（例: `NewFetcher`, `Article`）
- プライベート関数/変数: キャメルケース（例: `makeRootCmd`, `fetchClient`）
- インターフェース: 動詞または `-er` で終わる（例: `Fetcher`, `Viewer`）

## エラーハンドリング

- エラーは `fmt.Errorf("failed to ...: %w", err)` でコンテキスト付きラップ
- 複数エラーは `errors.Join()` で集約
- Sentinel Error は `errors.Is()` で判定

## レイヤード・アーキテクチャ

```
cmd/        → Presentation Layer（CLI定義のみ）
internal/
  app/      → Application Layer（ユースケース）
  domain/   → Domain Layer（ビジネスルール、インターフェース定義）
  infra/    → Infrastructure Layer（外部連携実装）
```

依存方向: `cmd → app → domain ← infra`（依存性逆転）

## 主要なパターン

- コンストラクタで依存性注入（`NewXxx(deps) *Xxx`）
- `context.Context` は関数の最初の引数に配置
- 早期リターンでネストを減らす
- ファイル末尾は必ず改行で終わる

## コミット前チェック

```bash
make fmt    # フォーマット
make lint   # 静的解析
make test   # テスト実行
```
