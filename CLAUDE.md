# CLAUDE.md

このファイルは、Claude Code (claude.ai/code) がこのリポジトリで作業する際のガイダンスを提供します。

## ビルド・開発コマンド

### 必須コマンド
```bash
# 開発依存関係のインストール（最初に実行）
make setup

# アプリケーションの実行
make run option="recommend"

# バイナリのビルド
make build

# テストの実行
make test

# コードの静的解析（重要：コミット前に必ず実行）
make lint

# コードのフォーマット
make fmt

# テスト用モックの生成
make generate

# ビルド成果物のクリーンアップ
make clean
```

### 特定のテストの実行
```bash
# 特定のパッケージのテストを実行
go test ./internal/domain/...

# 詳細な出力でテストを実行
go test -v ./...
```

## アーキテクチャ概要

アプリケーションの詳細なアーキテクチャについては [docs/02_architecture.md](docs/02_architecture.md) を参照してください。

## コーディングルール

詳細なコーディングルールは [docs/01_coding_rules.md](docs/01_coding_rules.md) を参照してください。

### 重要なポイント
- **コメントは全て日本語**で記述する
- コミット前に必ず `make lint` と `make fmt` を実行する
- 新しい外部依存関係を追加する際は `go mod tidy` を実行する
- インターフェースを変更した後は `make generate` でモックを再生成する

## セキュリティとトークン管理

### GitHub Personal Access Token の設定

GitHub CLIを使用するには、GH_TOKENが必要です。

#### トークンの作成
1. GitHub Settings > Developer settings > Personal access tokens > Tokens (classic)
2. 以下の権限を付与:
   - `repo` (リポジトリへの完全アクセス)
   - `read:org` (組織情報の読み取り)

#### 安全な設定方法
```bash
# プロジェクトルートに .env ファイルを作成（.gitignoreに含まれているため安全）
echo "GH_TOKEN=your_token_here" > .env
```

#### 重要な注意事項
- **絶対にトークンをコードやコミットに含めないでください**
- .envファイルは.gitignoreに含まれており、リポジトリにコミットされません
- 開発環境でのみ使用し、本番環境では適切な秘密管理システムを使用してください
- トークンは定期的にローテーションすることを推奨します

## その他
- リポジトリ情報は `git remote -v` で取得する
- 常に日本語で回答すること
- ファイル編集時には必ずファイル末尾が改行となるようにすること
- 対応の元になった github issue がある場合、プルリクエストの説明文には `fixed <issue番号>` を記載すること
    - (例) github issue #1 がある場合、 `fixed #1` を記載する
- プルリクエストのタイトルに、対応元の github issue 番号を含めないこと
