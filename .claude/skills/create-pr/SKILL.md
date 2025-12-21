---
name: create-pr
description: コミット前チェックを実行し、プルリクエストを作成します。現在のブランチからmainブランチへのPRを作成します。引数でPRタイトルと本文を受け取ります。
---

## 使用方法

### 前提条件
- コードの変更とコミットが完了していること
- mainブランチ以外のブランチにいること
- coding-specialist、architecture-specialist、testing-specialistでのレビューが完了していることを推奨
- **document-specialistでPR説明文を生成済みであること**

### スクリプトの実行

```bash
./.claude/skills/create-pr/create-pr.sh <PR_TITLE> <PR_BODY> [--dry-run]
```

### 引数

- `<PR_TITLE>`: PRのタイトル（必須）
  - issue番号を含めない（例: ○「ユーザー認証機能を追加」 ×「#123 ユーザー認証機能を追加」）
- `<PR_BODY>`: PRの本文（必須）
  - 改行を含む場合はヒアドキュメントを使用
  - `fixed #<issue番号>`を含める
- `--dry-run`: オプション。実際のプッシュとPR作成を行わず、実行内容のみ表示

### 実行手順

このスキルは以下の手順を自動実行します:

1. **引数の検証**
   - タイトルと本文が指定されていることを確認

2. **現在のブランチ確認**
   - mainブランチでの実行を防止

3. **コミット前チェック**
   - `make fmt`: コードフォーマット
   - `make lint`: 静的解析
   - `make test`: テスト実行

4. **変更のプッシュ**
   - プッシュが必要な場合、自動的に実行

5. **PR作成**
   - `gh pr create`で引数のタイトルと本文を使用してPRを作成
   - PR URLを表示

### 使用例

#### document-specialistと組み合わせた使用（推奨）

```bash
# 1. document-specialistでPR説明文を生成（Claude Codeコンテキスト内）
# 2. 生成された内容を使用してPR作成
./.claude/skills/create-pr/create-pr.sh "ユーザー認証機能を追加" "$(cat <<'EOF'
## 概要
ユーザー認証機能を実装しました。

## 変更内容
- `internal/app/auth/`: 認証サービスの実装
- `internal/infra/db/`: ユーザーテーブルの追加

## 関連Issue
fixed #123
EOF
)"
```

#### ドライラン（確認のみ）

```bash
./.claude/skills/create-pr/create-pr.sh "タイトル" "本文" --dry-run
```

### 注意事項

- **PRタイトルにissue番号を含めない**: GitHubの自動リンク機能を活用
- **関連issueは`fixed #番号`で記載**: PRマージ時にissueが自動クローズされます
- **テストが失敗した場合**: PR作成を中止し、修正してから再実行してください
- **document-specialistの使用**: PR説明文は必ずdocument-specialistで生成してから使用してください

### 詳細なガイドライン

PR作成の詳細なガイドライン、ブランチ戦略、コミット規約については、
[コントリビューションガイド](../../../docs/04_contributing.md)を参照してください。

### 関連エージェント・スキル

- **document-specialist エージェント**: PR説明文の生成（このスキル実行前に使用）
- coding-specialist エージェント: コード変更後のレビューに推奨
- architecture-specialist エージェント: アーキテクチャ変更時のレビューに推奨
- testing-specialist エージェント: テスト追加時のレビューに推奨
