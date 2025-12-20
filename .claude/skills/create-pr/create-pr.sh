#!/bin/bash
# プルリクエストを作成するスクリプト
#
# 使用方法: ./create-pr.sh <PR_TITLE> <PR_BODY>
#
# 注意事項:
#   - タイトルと本文は必須です
#   - PRタイトルにはissue番号を含めないでください
#   - 本文には fixed #<issue番号> を含めてください

set -euo pipefail

# 必要なコマンドの存在確認
for cmd in gh git make; do
    if ! command -v "$cmd" &> /dev/null; then
        echo "エラー: $cmd コマンドが見つかりません。インストールしてください。" >&2
        exit 1
    fi
done

# 引数チェック
if [[ $# -lt 2 ]]; then
    echo "使用方法: $0 <PR_TITLE> <PR_BODY>" >&2
    exit 1
fi

PR_TITLE="$1"
PR_BODY="$2"

if [[ -z "$PR_TITLE" ]]; then
    echo "エラー: PRタイトルが空です。" >&2
    exit 1
fi

if [[ -z "$PR_BODY" ]]; then
    echo "エラー: PR本文が空です。" >&2
    exit 1
fi

# 現在のブランチを取得
CURRENT_BRANCH=$(git branch --show-current)

if [[ "$CURRENT_BRANCH" == "main" ]]; then
    echo "エラー: mainブランチから直接PRを作成することはできません。" >&2
    exit 1
fi

echo "ブランチ: $CURRENT_BRANCH" >&2

# コミット前チェック
echo "コードフォーマット中..." >&2
if ! make fmt; then
    echo "エラー: コードフォーマットに失敗しました。" >&2
    exit 1
fi

if ! git diff --quiet; then
    echo "エラー: 'make fmt' によってファイルが変更されました。変更をコミットしてください。" >&2
    git status --short >&2
    exit 1
fi

echo "静的解析中..." >&2
if ! make lint; then
    echo "エラー: 静的解析でエラーが検出されました。" >&2
    exit 1
fi

echo "テスト実行中..." >&2
if ! make test; then
    echo "エラー: テストが失敗しました。" >&2
    exit 1
fi

# プッシュ処理
if git rev-parse --verify "origin/$CURRENT_BRANCH" &>/dev/null; then
    LOCAL_COMMIT=$(git rev-parse HEAD)
    REMOTE_COMMIT=$(git rev-parse "origin/$CURRENT_BRANCH")

    if [[ "$LOCAL_COMMIT" != "$REMOTE_COMMIT" ]]; then
        echo "プッシュ中..." >&2
        if ! git push origin "$CURRENT_BRANCH"; then
            echo "エラー: プッシュに失敗しました。" >&2
            exit 1
        fi
    fi
else
    echo "プッシュ中..." >&2
    if ! git push -u origin "$CURRENT_BRANCH"; then
        echo "エラー: プッシュに失敗しました。" >&2
        exit 1
    fi
fi

# PR作成
echo "PR作成中..." >&2

set +e
PR_URL=$(gh pr create \
    --title "$PR_TITLE" \
    --body "$PR_BODY" \
    --base main \
    --head "$CURRENT_BRANCH" 2>&1)
PR_EXIT_CODE=$?
set -e

if [ $PR_EXIT_CODE -ne 0 ]; then
    echo "エラー: プルリクエストの作成に失敗しました。" >&2
    echo "$PR_URL" >&2
    exit 1
fi

echo "PR URL: $PR_URL" >&2
exit 0
