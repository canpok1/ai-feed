# CLAUDE.md

このファイルは、Claude Code (claude.ai/code) がこのリポジトリで作業する際のガイダンスを提供します。

## ビルド・開発コマンド

詳細なビルド・開発コマンドについては [docs/00_development_setup.md](docs/00_development_setup.md) を参照してください。

### クイックリファレンス

**開発コマンド**:
```bash
make setup      # 開発環境セットアップ
make test       # テスト実行
make lint       # 静的解析（コミット前に必ず実行）
make fmt        # フォーマット
make build      # ビルド
make generate   # モック生成
```

**便利なスクリプト** (詳細は [docs/06_scripts.md](docs/06_scripts.md) を参照):
```bash
# バージョンタグの作成
./scripts/create-version-tag.sh [--dry-run]

# PR未解決コメントの確認
./scripts/get-pr-review-comments.sh <PR番号>

# レビュースレッドの解決
./scripts/resolve-review-thread.sh <スレッドID>
```

## アーキテクチャ概要

アプリケーションの詳細なアーキテクチャについては [docs/02_architecture_rules.md](docs/02_architecture_rules.md) を参照してください。

## コーディングルール

詳細なコーディングルールは [docs/01_coding_rules.md](docs/01_coding_rules.md) を参照してください。

プルリクエスト作成に関する詳細は [docs/04_contributing.md](docs/04_contributing.md) を参照してください。

## Claude Code 向け技術ガイダンス

### 作業開始前の確認
- **作業開始前に必ずdocs配下のドキュメントを確認してプロジェクトのルールを把握すること**
- 特に重要なドキュメント:
  - [docs/01_coding_rules.md](docs/01_coding_rules.md): コーディング規約
  - [docs/02_architecture_rules.md](docs/02_architecture_rules.md): アーキテクチャとディレクトリ構成
  - [docs/03_testing_rules.md](docs/03_testing_rules.md): テストの書き方
  - [docs/04_contributing.md](docs/04_contributing.md): プルリクエストの作成方法
- 各タスクに関連するドキュメントを事前に読み、ルールに従って作業を進めること

### コードレビュープロセス
- **ソースコード修正後は必ずレビューエージェントで自己レビューを実施すること**
- 使用するレビューエージェント:
  - `coding-rules-reviewer`: コーディングルール（docs/01_coding_rules.md）への準拠を確認
  - `architecture-rules-reviewer`: アーキテクチャルール（docs/02_architecture_rules.md）への準拠を確認
  - `testing-rules-reviewer`: テストルール（docs/03_testing_rules.md）への準拠を確認
- レビュー対象となる修正:
  - 新機能の実装
  - バグ修正
  - リファクタリング
  - その他、ソースコードを変更した場合
- レビュー後の指摘事項は必ず修正してからコミットすること

### 文書化プロセス
- **プルリクエスト作成時は必ずsummaristエージェントで作業内容を要約すること**
- **GitHub issue作成時も必ずsummaristエージェントで内容を整理すること**
- summaristが生成する内容:
  - プルリクエスト: .github/pull_request_template.md準拠の説明文
  - GitHub issue: 問題・詳細・期待される結果を含む説明文
  - 作業記録: シンプルで明確な要約
- 生成された内容を確認し、必要に応じて調整してから使用すること

### リポジトリ情報の取得
- リポジトリ情報は `git remote -v` で取得する

### シェルスクリプトのベストプラクティス
- 効率的なパラメータ展開を推奨
  - 例: `OWNER=${OWNER_REPO%/*}; REPO=${OWNER_REPO#*/}` （cutコマンドより効率的）

### コミュニケーション
- 常に日本語で回答すること
