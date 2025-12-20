#!/bin/bash
# PRレビュースレッドに返信を投稿するスクリプト
#
# 使用方法:
#   echo "コメント内容" | ./scripts/reply-to-review-thread.sh <スレッドID>
#   ./scripts/reply-to-review-thread.sh <スレッドID> <<EOF
#   複数行の
#   コメント内容
#   EOF
#
# 例:
#   echo "ご指摘ありがとうございます。修正しました。" | ./scripts/reply-to-review-thread.sh "xxxxxxxxxxxxxxxxxxxx"
#
# 注意事項:
#   - スレッドIDはGitHub GraphQL APIのNode ID形式で指定してください
#   - コメント投稿者への@メンションが自動的に追加されます
#   - コメント本文は標準入力から読み取ります

set -euo pipefail

# 必要なコマンドの存在確認
for cmd in gh jq; do
    if ! command -v "$cmd" &> /dev/null; then
        echo "エラー: $cmd コマンドが見つかりません。インストールしてください。" >&2
        exit 1
    fi
done

# 使用方法を表示
usage() {
    echo "使用方法: echo \"コメント内容\" | $0 <スレッドID>" >&2
    echo "" >&2
    echo "例:" >&2
    echo "  echo \"ご指摘ありがとうございます。\" | $0 \"xxxxxxxxxxxxxxxxxxxx\"" >&2
    echo "  $0 \"xxxxxxxxxxxxxxxxxxxx\" <<EOF" >&2
    echo "  複数行のコメント" >&2
    echo "  EOF" >&2
    exit 1
}

# 引数チェック
if [[ $# -ne 1 ]]; then
    echo "エラー: 引数が不正です。" >&2
    usage
fi

THREAD_ID="$1"

# スレッドIDが空でないかチェック
if [[ -z "$THREAD_ID" ]]; then
    echo "エラー: スレッドIDが空です。" >&2
    exit 1
fi

# 標準入力からコメント本文を読み取り
echo "コメント本文を読み取り中..." >&2
COMMENT_BODY=$(cat)

# コメント本文が空でないかチェック
if [[ -z "$COMMENT_BODY" ]]; then
    echo "エラー: コメント本文が空です。" >&2
    exit 1
fi

# リポジトリ情報を取得
echo "リポジトリ情報を取得中..." >&2
OWNER_REPO=$(gh repo view --json nameWithOwner --jq '.nameWithOwner')
OWNER="${OWNER_REPO%/*}"
REPO="${OWNER_REPO#*/}"

echo "リポジトリ: $OWNER/$REPO" >&2
echo "スレッドID: $THREAD_ID" >&2
echo "" >&2

# スレッド情報を取得（コメント投稿者のユーザー名を含む）
echo "スレッド情報を取得中..." >&2

set +e
THREAD_INFO=$(gh api graphql \
  -f threadId="$THREAD_ID" \
  -f query='
query($threadId: ID!) {
  node(id: $threadId) {
    ... on PullRequestReviewThread {
      id
      comments(last: 1) {
        nodes {
          author {
            login
          }
        }
      }
    }
  }
}' 2>&1)
GH_EXIT_CODE=$?
set -e

# gh api コマンドがエラーの場合
if [ $GH_EXIT_CODE -ne 0 ]; then
    echo "エラー: スレッドID '$THREAD_ID' が見つからないか、アクセス権がありません。" >&2
    exit 1
fi

# スレッドが存在しない場合（nullの場合）
if echo "$THREAD_INFO" | jq -e '.data.node == null' > /dev/null 2>&1; then
    echo "エラー: スレッドID '$THREAD_ID' が見つからないか、アクセス権がありません。" >&2
    exit 1
fi

# コメント投稿者のユーザー名を取得
AUTHOR_LOGIN=$(echo "$THREAD_INFO" | jq -r '.data.node.comments.nodes[-1].author.login // empty')

if [[ -z "$AUTHOR_LOGIN" ]]; then
    echo "エラー: コメント投稿者の情報を取得できませんでした。" >&2
    exit 1
fi

echo "コメント投稿者: @$AUTHOR_LOGIN" >&2

# コメント本文の先頭にメンションを追加
COMMENT_WITH_MENTION="@$AUTHOR_LOGIN ${COMMENT_BODY}"

echo "" >&2
echo "返信を投稿中..." >&2

# 返信を投稿
set +e
RESULT=$(gh api graphql \
  -f pullRequestReviewThreadId="$THREAD_ID" \
  -f body="$COMMENT_WITH_MENTION" \
  -f query='
mutation($pullRequestReviewThreadId: ID!, $body: String!) {
  addPullRequestReviewThreadReply(input: {pullRequestReviewThreadId: $pullRequestReviewThreadId, body: $body}) {
    comment {
      id
      body
      createdAt
      author {
        login
      }
    }
  }
}' 2>&1)
GH_POST_EXIT_CODE=$?
set -e

# 結果を確認
if [ "$GH_POST_EXIT_CODE" -ne 0 ]; then
    echo "" >&2
    echo "✗ 返信の投稿に失敗しました。" >&2
    echo "エラー詳細:" >&2
    echo "$RESULT" >&2
    exit 1
fi

COMMENT_ID=$(echo "$RESULT" | jq -r '.data.addPullRequestReviewThreadReply.comment.id // empty')

if [[ -n "$COMMENT_ID" ]]; then
    echo "" >&2
    echo "✓ 返信を投稿しました。" >&2
    echo "$RESULT" | jq -r '.data.addPullRequestReviewThreadReply.comment | "コメントID: \(.id)\n投稿者: @\(.author.login)\n作成日時: \(.createdAt)"' >&2
    exit 0
else
    echo "" >&2
    echo "✗ 返信の投稿に失敗しました。" >&2
    echo "エラー詳細:" >&2
    echo "$RESULT" | jq >&2
    exit 1
fi
