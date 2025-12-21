---
name: reply-to-review-thread
description: 指定したレビュースレッドに返信コメントを投稿します。コメント投稿者への@メンションが自動的に追加され、標準入力からコメント本文を読み取ります。
---

## 使用方法

### スクリプトの実行

```bash
# パイプ経由でコメントを投稿
echo "コメント内容" | ./.claude/skills/reply-to-review-thread/reply-to-review-thread.sh <スレッドID>

# ヒアドキュメントで複数行コメントを投稿
./.claude/skills/reply-to-review-thread/reply-to-review-thread.sh <スレッドID> <<EOF
複数行の
コメント内容
EOF
```

### 引数

- `<スレッドID>`: 返信したいレビュースレッドのID（例: `xxxxxxxxxxxxxxxxxxxx`）
  - スレッドIDは引用符で囲んで指定することを推奨
- コメント本文: 標準入力から読み取り

### 出力例

```bash
コメント本文を読み取り中...
リポジトリ情報を取得中...
リポジトリ: owner/ai-feed
スレッドID: xxxxxxxxxxxxxxxxxxxx

スレッド情報を取得中...
コメント投稿者: @reviewer-username

返信を投稿中...

✓ 返信を投稿しました。
コメントID: PRRC_kwDONTZR484BhKaJ
投稿者: @your-username
作成日時: 2025-12-13T10:30:00Z
```

### 使用例

```bash
# シンプルな返信
echo "ご指摘ありがとうございます。修正しました。" | ./.claude/skills/reply-to-review-thread/reply-to-review-thread.sh "xxxxxxxxxxxxxxxxxxxx"

# 複数行の詳細な返信
./.claude/skills/reply-to-review-thread/reply-to-review-thread.sh "xxxxxxxxxxxxxxxxxxxx" <<EOF
ご指摘ありがとうございます。

以下の対応を行いました：
- エラーハンドリングを改善
- テストケースを追加
EOF
```

### 注意事項

- スレッドにコメントを投稿する権限が必要です
- コメント投稿者への@メンションは自動的に追加されるため、手動で記述する必要はありません
- 標準入力からコメントを読み取るため、空のコメントは投稿できません
- スレッドIDはGitHub GraphQL APIのNode ID形式（例: `PRRT_kwDO...`）で指定
