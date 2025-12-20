#!/bin/bash
# プルリクエストを作成するスクリプト
#
# 使用方法:
#   ./create-pr.sh <PR_TITLE> <PR_BODY> [--dry-run]
#
# 例:
#   ./create-pr.sh "ユーザー認証機能を追加" "$(cat <<'EOF'
#   ## 概要
#   ユーザー認証機能を実装しました。
#
#   ## 関連Issue
#   fixed #123
#   EOF
#   )"
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

# 使用方法を表示
usage() {
    echo "使用方法: $0 <PR_TITLE> <PR_BODY> [--dry-run]" >&2
    echo "" >&2
    echo "引数:" >&2
    echo "  PR_TITLE  PRのタイトル（issue番号を含めない）" >&2
    echo "  PR_BODY   PRの本文（fixed #<issue番号> を含める）" >&2
    echo "  --dry-run 実行内容のみ表示（実際の操作は行わない）" >&2
    echo "" >&2
    echo "例:" >&2
    echo '  $0 "機能を追加" "$(cat <<'"'"'EOF'"'"'' >&2
    echo '  ## 概要' >&2
    echo '  機能を追加しました。' >&2
    echo '  ' >&2
    echo '  ## 関連Issue' >&2
    echo '  fixed #123' >&2
    echo '  EOF' >&2
    echo '  )"' >&2
    exit 1
}

# 引数チェック
DRY_RUN=false
if [[ $# -lt 2 ]]; then
    echo "エラー: タイトルと本文を指定してください。" >&2
    usage
fi

PR_TITLE="$1"
PR_BODY="$2"

# 3番目の引数があれば--dry-runかチェック
if [[ $# -ge 3 ]]; then
    if [[ "$3" == "--dry-run" ]]; then
        DRY_RUN=true
        echo "ドライランモード: 実際の操作は行いません。" >&2
        echo "" >&2
    else
        echo "エラー: 不明なオプション '$3'" >&2
        usage
    fi
fi

# タイトルと本文が空でないかチェック
if [[ -z "$PR_TITLE" ]]; then
    echo "エラー: PRタイトルが空です。" >&2
    exit 1
fi

if [[ -z "$PR_BODY" ]]; then
    echo "エラー: PR本文が空です。" >&2
    exit 1
fi

echo "====================================================================" >&2
echo "プルリクエスト作成処理を開始します" >&2
echo "====================================================================" >&2
echo "" >&2

# 現在のブランチを取得
CURRENT_BRANCH=$(git branch --show-current)
echo "現在のブランチ: $CURRENT_BRANCH" >&2

# mainブランチでないことを確認
if [[ "$CURRENT_BRANCH" == "main" ]]; then
    echo "エラー: mainブランチから直接PRを作成することはできません。" >&2
    echo "フィーチャーブランチを作成してください。" >&2
    exit 1
fi

echo "" >&2
echo "--------------------------------------------------------------------" >&2
echo "ステップ 1/5: コミット前チェックを実行" >&2
echo "--------------------------------------------------------------------" >&2

# make fmt
echo "" >&2
echo "[1/3] コードフォーマット (make fmt)..." >&2
if $DRY_RUN; then
    echo "[ドライラン] make fmt をスキップ" >&2
else
    if ! make fmt; then
        echo "エラー: コードフォーマットに失敗しました。" >&2
        exit 1
    fi
    echo "✓ コードフォーマット完了" >&2
fi

# make lint
echo "" >&2
echo "[2/3] 静的解析 (make lint)..." >&2
if $DRY_RUN; then
    echo "[ドライラン] make lint をスキップ" >&2
else
    if ! make lint; then
        echo "エラー: 静的解析でエラーが検出されました。修正してください。" >&2
        exit 1
    fi
    echo "✓ 静的解析完了" >&2
fi

# make test
echo "" >&2
echo "[3/3] テスト実行 (make test)..." >&2
if $DRY_RUN; then
    echo "[ドライラン] make test をスキップ" >&2
else
    if ! make test; then
        echo "エラー: テストが失敗しました。修正してください。" >&2
        exit 1
    fi
    echo "✓ テスト完了" >&2
fi

echo "" >&2
echo "--------------------------------------------------------------------" >&2
echo "ステップ 2/5: 変更のプッシュ" >&2
echo "--------------------------------------------------------------------" >&2

# リモートブランチとの差分を確認
if git rev-parse --verify "origin/$CURRENT_BRANCH" &>/dev/null; then
    # リモートブランチが存在する場合
    LOCAL_COMMIT=$(git rev-parse HEAD)
    REMOTE_COMMIT=$(git rev-parse "origin/$CURRENT_BRANCH")

    if [[ "$LOCAL_COMMIT" != "$REMOTE_COMMIT" ]]; then
        echo "ローカルとリモートに差分があります。プッシュが必要です。" >&2
        if $DRY_RUN; then
            echo "[ドライラン] git push origin $CURRENT_BRANCH をスキップ" >&2
        else
            echo "変更をプッシュしています..." >&2
            if ! git push origin "$CURRENT_BRANCH"; then
                echo "エラー: プッシュに失敗しました。" >&2
                exit 1
            fi
            echo "✓ プッシュ完了" >&2
        fi
    else
        echo "リモートブランチは最新です。" >&2
    fi
else
    # リモートブランチが存在しない場合
    echo "リモートブランチが存在しません。プッシュが必要です。" >&2
    if $DRY_RUN; then
        echo "[ドライラン] git push -u origin $CURRENT_BRANCH をスキップ" >&2
    else
        echo "変更をプッシュしています..." >&2
        if ! git push -u origin "$CURRENT_BRANCH"; then
            echo "エラー: プッシュに失敗しました。" >&2
            exit 1
        fi
        echo "✓ プッシュ完了" >&2
    fi
fi

echo "" >&2
echo "--------------------------------------------------------------------" >&2
echo "ステップ 3/5: リポジトリ情報の取得" >&2
echo "--------------------------------------------------------------------" >&2

OWNER_REPO=$(gh repo view --json nameWithOwner --jq '.nameWithOwner')
echo "リポジトリ: $OWNER_REPO" >&2

echo "" >&2
echo "--------------------------------------------------------------------" >&2
echo "ステップ 4/5: PR内容の確認" >&2
echo "--------------------------------------------------------------------" >&2

echo "PRタイトル: $PR_TITLE" >&2
echo "" >&2
echo "PR本文:" >&2
echo "$PR_BODY" >&2

echo "" >&2
echo "--------------------------------------------------------------------" >&2
echo "ステップ 5/5: プルリクエストの作成" >&2
echo "--------------------------------------------------------------------" >&2

if $DRY_RUN; then
    echo "[ドライラン] gh pr create をスキップ" >&2
    echo "" >&2
    echo "====================================================================" >&2
    echo "ドライラン完了" >&2
    echo "====================================================================" >&2
    exit 0
fi

# PR作成
echo "プルリクエストを作成しています..." >&2

set +e
PR_URL=$(gh pr create \
    --title "$PR_TITLE" \
    --body "$PR_BODY" \
    --base main \
    --head "$CURRENT_BRANCH" 2>&1)
PR_EXIT_CODE=$?
set -e

if [ $PR_EXIT_CODE -ne 0 ]; then
    echo "" >&2
    echo "✗ プルリクエストの作成に失敗しました。" >&2
    echo "エラー詳細:" >&2
    echo "$PR_URL" >&2
    exit 1
fi

echo "" >&2
echo "====================================================================" >&2
echo "✓ プルリクエストを作成しました！" >&2
echo "====================================================================" >&2
echo "" >&2
echo "PR URL: $PR_URL" >&2
echo "" >&2

exit 0
