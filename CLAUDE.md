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

# golangci-lintによる包括的な静的解析（重要：コミット前に必ず実行）
make lint

# テストカバレッジレポート生成と閾値チェック（60%以上、将来的に70%目標）
make test-coverage

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

## Claude Code エージェント

このプロジェクトでは、作業効率を向上させるための専用エージェントを使用できます。

### code-reviewer
コード実装完了後の品質レビューに使用します。

**使用タイミング**:
- 新機能実装完了後
- バグ修正完了後
- リファクタリング完了後

**提供される内容**:
- プロジェクト規約（日本語コメント、クリーンアーキテクチャ等）の遵守確認
- コード品質のチェック（エラーハンドリング、テスタビリティ等）
- セキュリティとベストプラクティスの確認
- 改善提案と具体的な修正案

### summarist
作業内容の要約・文書化に使用します。

**使用タイミング**:
- プルリクエスト作成時の説明文生成
- GitHub issue作成時の内容整理
- 作業記録の文書化

**提供される内容**:
- プルリクエスト用の説明文（.github/pull_request_template.md準拠）
- GitHub issue用の説明文（問題・詳細・期待される結果）
- シンプルな作業記録

**推奨ワークフロー**:
1. コード実装・修正完了
2. code-reviewerでコード品質チェック
3. 指摘事項を修正
4. summaristで作業内容を要約
5. プルリクエスト作成

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
- シェルスクリプトでは効率的なパラメータ展開を推奨
  - 例: `OWNER=${OWNER_REPO%/*}; REPO=${OWNER_REPO#*/}` （cutコマンドより効率的）
- 常に日本語で回答すること
- ファイル編集時には必ずファイル末尾が改行となるようにすること
- 対応の元になった github issue がある場合、プルリクエストの説明文には `fixed <issue番号>` を記載すること
    - (例) github issue #1 がある場合、 `fixed #1` を記載する
- プルリクエストのタイトルに、対応元の github issue 番号を含めないこと
