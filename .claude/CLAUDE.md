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
- **ソースコード修正後は必ず専門家エージェントで自己レビューを実施すること**
- 使用する専門家エージェント:
  - `coding-specialist`: コーディングルール（docs/01_coding_rules.md）に関する相談・レビュー・ガイダンス
  - `architecture-specialist`: アーキテクチャルール（docs/02_architecture_rules.md）に関する相談・レビュー・ガイダンス
  - `testing-specialist`: テストルール（docs/03_testing_rules.md）に関する相談・レビュー・ガイダンス
- エージェントの活用方法:
  - **実装前**: 設計や方針についてガイダンスを依頼
  - **実装中**: ルールの解釈や適用方法について相談
  - **実装後**: コードのレビューを依頼
- レビュー対象となる修正:
  - 新機能の実装
  - バグ修正
  - リファクタリング
  - その他、ソースコードを変更した場合
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

### コミュニケーション
- 常に日本語で回答すること
