#!/bin/bash
# GitHub issue情報を取得するスクリプト
#
# 使用方法:
#   ./plan-issue.sh <ISSUE_NUMBER>
#
# 例:
#   ./plan-issue.sh "#123"
#   ./plan-issue.sh "123"
#
# 注意事項:
#   - issue番号は必須です（#付きでも無しでも可）
#   - 取得した情報を元にClaude Codeがタスクファイルを作成します

set -euo pipefail

# 必要なコマンドの存在確認
for cmd in gh git; do
    if ! command -v "$cmd" &> /dev/null; then
        echo "エラー: $cmd コマンドが見つかりません。インストールしてください。" >&2
        exit 1
    fi
done

# 使用方法を表示
usage() {
    echo "使用方法: $0 <ISSUE_NUMBER>" >&2
    echo "" >&2
    echo "引数:" >&2
    echo "  ISSUE_NUMBER  GitHub issueの番号（#付きでも無しでも可）" >&2
    echo "" >&2
    echo "例:" >&2
    echo "  $0 \"#123\"" >&2
    echo "  $0 \"123\"" >&2
    exit 1
}

# 引数チェック
if [[ $# -lt 1 ]]; then
    echo "エラー: issue番号を指定してください。" >&2
    usage
fi

ISSUE_NUMBER="$1"

# issue番号から#を除去（あれば）
ISSUE_NUMBER="${ISSUE_NUMBER#\#}"

# 数字かどうかチェック
if ! [[ "$ISSUE_NUMBER" =~ ^[0-9]+$ ]]; then
    echo "エラー: 無効なissue番号です: $ISSUE_NUMBER" >&2
    exit 1
fi

echo "====================================================================" >&2
echo "GitHub Issue情報の取得" >&2
echo "====================================================================" >&2
echo "" >&2

# リポジトリ情報の取得
OWNER_REPO=$(git remote -v | grep -m1 origin | sed 's/.*github\.com[:/]\(.*\)\.git.*/\1/')
echo "リポジトリ: $OWNER_REPO" >&2
echo "Issue番号: #$ISSUE_NUMBER" >&2
echo "" >&2

# issue情報を取得
echo "--------------------------------------------------------------------" >&2
echo "Issue情報を取得中..." >&2
echo "--------------------------------------------------------------------" >&2
echo "" >&2

set +e
ISSUE_JSON=$(gh issue view "$ISSUE_NUMBER" --json number,title,body,state,labels,assignees,milestone,createdAt,updatedAt 2>&1)
ISSUE_EXIT_CODE=$?
set -e

if [ $ISSUE_EXIT_CODE -ne 0 ]; then
    echo "エラー: issue情報の取得に失敗しました。" >&2
    echo "詳細: $ISSUE_JSON" >&2
    exit 1
fi

# JSON情報を整形して表示
echo "====================================================================" >&2
echo "Issue #$ISSUE_NUMBER の情報" >&2
echo "====================================================================" >&2
echo "" >&2

# タイトル
TITLE=$(echo "$ISSUE_JSON" | jq -r '.title')
echo "【タイトル】" >&2
echo "$TITLE" >&2
echo "" >&2

# 状態
STATE=$(echo "$ISSUE_JSON" | jq -r '.state')
echo "【状態】" >&2
echo "$STATE" >&2
echo "" >&2

# 本文
BODY=$(echo "$ISSUE_JSON" | jq -r '.body // "（本文なし）"')
echo "【本文】" >&2
echo "$BODY" >&2
echo "" >&2

# ラベル
LABELS=$(echo "$ISSUE_JSON" | jq -r '.labels[]?.name // empty' | tr '\n' ', ' | sed 's/,$//')
if [[ -n "$LABELS" ]]; then
    echo "【ラベル】" >&2
    echo "$LABELS" >&2
    echo "" >&2
fi

# アサイニー
ASSIGNEES=$(echo "$ISSUE_JSON" | jq -r '.assignees[]?.login // empty' | tr '\n' ', ' | sed 's/,$//')
if [[ -n "$ASSIGNEES" ]]; then
    echo "【担当者】" >&2
    echo "$ASSIGNEES" >&2
    echo "" >&2
fi

# マイルストーン
MILESTONE=$(echo "$ISSUE_JSON" | jq -r '.milestone.title // empty')
if [[ -n "$MILESTONE" ]]; then
    echo "【マイルストーン】" >&2
    echo "$MILESTONE" >&2
    echo "" >&2
fi

# 作成日時・更新日時
CREATED_AT=$(echo "$ISSUE_JSON" | jq -r '.createdAt')
UPDATED_AT=$(echo "$ISSUE_JSON" | jq -r '.updatedAt')
echo "【作成日時】" >&2
echo "$CREATED_AT" >&2
echo "" >&2
echo "【更新日時】" >&2
echo "$UPDATED_AT" >&2
echo "" >&2

echo "====================================================================" >&2
echo "情報取得完了" >&2
echo "====================================================================" >&2
echo "" >&2
echo "この情報を元に、tmp/todo フォルダにタスクファイルを作成してください。" >&2

exit 0
