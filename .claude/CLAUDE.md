# CLAUDE.md

このファイルは、Claude Code (claude.ai/code) がこのリポジトリで作業する際のガイダンスを提供します。

## ビルド・開発コマンド

詳細なビルド・開発コマンドについては [docs/00_development_setup.md](../docs/00_development_setup.md) を参照してください。

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

## アーキテクチャ概要

アプリケーションの詳細なアーキテクチャについては [docs/02_architecture_rules.md](../docs/02_architecture_rules.md) を参照してください。

## コーディングルール

詳細なコーディングルールは [docs/01_coding_rules.md](../docs/01_coding_rules.md) を参照してください。

プルリクエスト作成に関する詳細は [docs/04_contributing.md](../docs/04_contributing.md) を参照してください。

## Claude Code 向け技術ガイダンス

### 作業開始前の確認
- 詳細なルールは `.claude/rules/` およびdocs/配下のドキュメントを参照すること

### コードレビュープロセス
- ソースコード修正後は専門家エージェントで自己レビューを実施すること
- 詳細な手順は `review` スキルを参照
- レビュー後の指摘事項は必ず修正してからコミットすること

### 文書化プロセス
- **プルリクエスト作成時は必ずdocument-specialistエージェントで作業内容を要約すること**
- **GitHub issue作成時も必ずdocument-specialistエージェントで内容を整理すること**
- 詳細な文書化ルールは [docs/07_document_rules.md](../docs/07_document_rules.md) を参照
- document-specialistが生成する内容:
  - プルリクエスト: .github/pull_request_template.md準拠の説明文
  - GitHub issue: 問題・詳細・期待される結果を含む説明文
  - 作業記録: シンプルで明確な要約
- 生成された内容を確認し、必要に応じて調整してから使用すること

### リポジトリ情報の取得
- リポジトリ情報は `git remote -v` で取得する

### シェルスクリプトのベストプラクティス
- 効率的なパラメータ展開を推奨
  - 例: `OWNER=${OWNER_REPO%/*}; REPO=${OWNER_REPO#*/}` （cutコマンドより効率的）

### ユーザーへの確認
- 判断が必要な場面では `AskUserQuestion` ツールを使用してユーザーに確認すること
- 曖昧な要件や複数の実装方法がある場合は、推測せず必ず確認する

### コミュニケーション
- 常に日本語で回答すること
