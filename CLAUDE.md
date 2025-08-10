# CLAUDE.md

このファイルは、Claude Code (claude.ai/code) がこのリポジトリで作業する際のガイダンスを提供します。

## ビルド・開発コマンド

### 必須コマンド
```bash
# 開発依存関係のインストール（最初に実行）
make setup

# アプリケーションの実行
make run option="preview"
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
# 特定のテストファイルを実行
go test ./cmd/preview_test.go

# 特定のパッケージのテストを実行
go test ./internal/domain/...

# 詳細な出力でテストを実行
go test -v ./...
```

## アーキテクチャ概要

このプロジェクトはCobraフレームワークを使用したGo CLIアプリケーションで、クリーンアーキテクチャパターンを採用しています：

- **cmd/**: Cobraを使用したCLIコマンドの実装。各コマンド（preview、recommend、config、profile）は個別のファイルに分かれています。
- **internal/domain/**: インターフェースとエンティティを含むコアビジネスロジック。アプリケーションのメインロジックが存在します。
- **internal/infra/**: 外部サービス（RSSフェッチ、AI連携、Slack/Misskey投稿）のインフラストラクチャ実装。

依存性注入パターンに従い、コマンドはコンストラクタを通じて依存関係を受け取るため、モックを使用したテストが可能です。

## 主要なアーキテクチャ上の決定事項

1. **設定システム**: YAMLファイル（config.ymlとprofile.yml）を使用。configコマンドでこれらのファイルを管理します。

2. **AI連携**: 現在はGoogle Gemini APIを使用して記事の推薦文を生成。AIのプロンプトと動作はprofile.ymlでカスタマイズ可能です。

3. **テスト戦略**: testifyフレームワークを使用したテーブル駆動テスト。モックはgo.uber.org/mockで生成されます。

4. **外部連携**: SlackとMisskeyプラットフォームへの投稿をサポート。各連携はinternal/infra/内の個別のviewerとして実装されています。

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
