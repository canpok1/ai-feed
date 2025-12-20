---
name: update-pr
description: 既存のプルリクエストの説明文を更新します。document-specialistで生成した内容でPRを更新する際に使用します。
---

# update-pr スキル

既存のプルリクエストのタイトルや説明文を更新するスキルです。

## 使用方法

```bash
./.claude/skills/update-pr/update-pr.sh <PR番号> [--title "新しいタイトル"] [--dry-run]
```

PR本文は標準入力から受け取ります（ヒアドキュメントやパイプ経由）。

## 引数

### 必須引数

- `<PR番号>`: 更新対象のプルリクエスト番号（例: `123` または `#123`）

### オプション引数

- `--title "新しいタイトル"`: PRのタイトルを更新する場合に指定
- `--dry-run`: 実際の更新を行わず、変更内容をプレビュー表示のみ行う

### PR本文の指定方法

PR本文は標準入力から読み取ります。以下の方法で指定できます：

**ヒアドキュメントを使用:**
```bash
./.claude/skills/update-pr/update-pr.sh 123 --title "新タイトル" <<'EOF'
## Summary
- 変更内容1
- 変更内容2

## Test plan
- [ ] テスト項目1
- [ ] テスト項目2

🤖 Generated with [Claude Code](https://claude.com/claude-code)
EOF
```

**パイプを使用:**
```bash
echo "PR本文" | ./.claude/skills/update-pr/update-pr.sh 123
```

**ファイルから読み込み:**
```bash
cat pr_body.md | ./.claude/skills/update-pr/update-pr.sh 123 --title "新タイトル"
```

## 使用例

### 1. タイトルのみ更新

```bash
echo "" | ./.claude/skills/update-pr/update-pr.sh 123 --title "fix: バグ修正のタイトル"
```

本文を更新しない場合でも、標準入力は必要です（空文字列でも可）。

### 2. 本文のみ更新

```bash
./.claude/skills/update-pr/update-pr.sh 123 <<'EOF'
## Summary
- データベース接続の改善
- エラーハンドリングの追加

## Test plan
- [x] ユニットテスト追加
- [x] 統合テスト実行

🤖 Generated with [Claude Code](https://claude.com/claude-code)
EOF
```

### 3. タイトルと本文の両方を更新

```bash
./.claude/skills/update-pr/update-pr.sh 123 --title "feat: 新機能の追加" <<'EOF'
## Summary
- ユーザー認証機能の追加
- セッション管理の実装

## Test plan
- [x] 認証フローのテスト
- [x] セッション管理のテスト

🤖 Generated with [Claude Code](https://claude.com/claude-code)
EOF
```

### 4. ドライラン（プレビュー）

```bash
./.claude/skills/update-pr/update-pr.sh 123 --title "新タイトル" --dry-run <<'EOF'
新しいPR本文
EOF
```

実際の更新は行わず、変更内容のみを表示します。

## 出力

スクリプトは以下の情報を stderr に出力します：

- 処理状況（PR確認中、更新中など）
- エラーメッセージ（存在しない場合）
- 更新完了メッセージ

実際のPRデータは stdout には出力されません。

## 典型的なワークフロー

### document-specialist エージェントとの連携

1. **作業内容の完了**: コード修正やリファクタリングを完了
2. **PR作成**: `create-pr` スキルで初期PRを作成
3. **レビューコメント対応**: レビューで指摘された内容を修正
4. **PR説明文の更新**:
   - document-specialist エージェントを起動して、追加変更を含む更新されたPR説明文を生成
   - 生成された説明文を `update-pr` スキルでPRに反映

**実行例:**
```
ユーザー: 「document-specialistエージェントを使ってPR #123の説明文を更新してください」

Claude: document-specialistエージェントを起動し、変更内容をまとめます...
[エージェントが作業内容を分析し、説明文を生成]

Claude: 生成されたPR説明文をPR #123に反映します。

./.claude/skills/update-pr/update-pr.sh 123 --title "更新されたタイトル" <<'EOF'
[生成されたPR説明文]
EOF
```

### 手動でのPR更新

レビューコメントへの対応後、手動でPR説明文を更新したい場合：

```bash
# 1. レビューコメントに対応したコードを修正
# 2. 変更内容を反映したPR説明文を準備
# 3. update-prスキルで更新

./.claude/skills/update-pr/update-pr.sh 456 <<'EOF'
## Summary
- レビューコメントの指摘に対応
  - エラーハンドリングの改善
  - テストカバレッジの向上

## Changes
- `pkg/service/auth.go`: エラーメッセージの詳細化
- `pkg/service/auth_test.go`: エッジケースのテスト追加

## Test plan
- [x] 既存テストが全て通過
- [x] 新規テストケース追加

🤖 Generated with [Claude Code](https://claude.com/claude-code)
EOF
```

## 注意事項

### 前提条件

- GitHub CLI (`gh`) がインストールされ、認証済みであること
- `jq` コマンドがインストールされていること
- 対象のプルリクエストが存在すること
- プルリクエストを編集する権限があること

### 制限事項

- PR番号は必須です
- タイトルまたは本文の少なくとも一方を指定する必要があります
  - タイトルのみ: `--title` オプションを使用し、標準入力は空でも可
  - 本文のみ: 標準入力から本文を渡す
  - 両方: `--title` オプションと標準入力の両方を使用
- `--dry-run` モードでは、実際の更新は行われません

### エラー処理

スクリプトは以下の場合にエラーを返します：

- 必要なコマンド（`gh`, `jq`）が見つからない場合
- PR番号が指定されていない場合
- 指定されたPRが存在しない場合
- PR更新権限がない場合
- GitHub APIエラーが発生した場合

## 関連スキル/エージェント

### 関連スキル

- **create-pr**: 新規プルリクエストの作成
  - PRを最初に作成する際に使用
  - document-specialistとの連携推奨
- **get-pr-review-comments**: PRのレビューコメント取得
  - レビューコメントを確認する際に使用
  - 対応すべき指摘事項の把握に役立つ
- **reply-to-review-thread**: レビュースレッドへの返信
  - レビューコメントに対して返信する際に使用

### 関連エージェント

- **document-specialist**: 作業内容の文書化専門エージェント
  - PR説明文の生成に特化
  - 変更内容を簡潔にまとめる
  - `.github/pull_request_template.md` 準拠の形式で出力
  - `update-pr` スキルと組み合わせて使用することを推奨

## 参考資料

- `.github/pull_request_template.md`: PRテンプレート
- `docs/07_document_rules.md`: 文書化ルール
- `docs/04_contributing.md`: コントリビューションガイド
