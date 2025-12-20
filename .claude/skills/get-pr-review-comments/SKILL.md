---
name: get-pr-review-comments
description: PRの未解決レビューコメントを取得します。
---

## 手順

このスキルは、指定したプルリクエストの未解決レビューコメントをJSON形式で取得します。

### スクリプトの実行

```bash
./.claude/skills/get-pr-review-comments/get-pr-review-comments.sh <PR番号>
```

### 引数

- `<PR番号>`: 確認したいプルリクエストの番号（数値のみ、#は不要）

### 出力形式

各未解決レビュースレッドについて、以下の情報をJSON形式（NDJSON）で出力します：

```json
{
  "thread_id": "スレッドID（例: PRRT_kwDONTZR484BhKaH）",
  "author": "コメント投稿者のGitHubユーザー名",
  "comment": "最新のコメント内容"
}
```

### 使用例

```bash
# PR #123の未解決コメントを取得
./.claude/skills/get-pr-review-comments/get-pr-review-comments.sh 123
```

### 次のアクション

取得した`thread_id`を使って：
- スレッドの詳細を確認: `get-pr-review-thread-details` スキル
- スレッドに返信: `reply-to-review-thread` スキル
- スレッドを解決: `resolve-pr-thread` スキル
