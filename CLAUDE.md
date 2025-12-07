# CLAUDE.md

このファイルは、Claude Code (claude.ai/code) がこのリポジトリで作業する際のガイダンスを提供します。

## ビルド・開発コマンド

詳細なビルド・開発コマンドについては [docs/00_development_setup.md](docs/00_development_setup.md) を参照してください。

### クイックリファレンス
```bash
make setup      # 開発環境セットアップ
make test       # テスト実行
make lint       # 静的解析（コミット前に必ず実行）
make fmt        # フォーマット
make build      # ビルド
make generate   # モック生成
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

プルリクエスト作成に関する詳細は [docs/04_contributing.md](docs/04_contributing.md) を参照してください。

## Claude Code 向け技術ガイダンス

### リポジトリ情報の取得
- リポジトリ情報は `git remote -v` で取得する

### シェルスクリプトのベストプラクティス
- 効率的なパラメータ展開を推奨
  - 例: `OWNER=${OWNER_REPO%/*}; REPO=${OWNER_REPO#*/}` （cutコマンドより効率的）

### コミュニケーション
- 常に日本語で回答すること
