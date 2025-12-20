#!/bin/bash
# プルリクエストの内容を更新するスクリプト
#
# 使用方法:
#   ./update-pr.sh <PR番号> [--title "新しいタイトル"] [--dry-run]
#
# PR本文は標準入力から読み取ります。
#
# 例:
#   # 本文のみ更新
#   ./update-pr.sh 123 <<'EOF'
#   新しいPR本文
#   EOF
#
#   # タイトルと本文を更新
#   ./update-pr.sh 123 --title "新しいタイトル" <<'EOF'
#   新しいPR本文
#   EOF
#
#   # ドライラン
#   ./update-pr.sh 123 --title "新しいタイトル" --dry-run <<'EOF'
#   新しいPR本文
#   EOF

set -euo pipefail

# 必要なコマンドの存在確認
for cmd in gh jq; do
    if ! command -v "$cmd" &> /dev/null; then
        echo "エラー: $cmd コマンドが見つかりません。" >&2
        echo "GitHub CLI (gh) と jq をインストールしてください。" >&2
        exit 1
    fi
done

# usage関数
usage() {
    echo "使用方法: $0 <PR番号> [--title \"新しいタイトル\"] [--dry-run]" >&2
    echo "" >&2
    echo "引数:" >&2
    echo "  <PR番号>    更新対象のプルリクエスト番号（例: 123 または #123）" >&2
    echo "" >&2
    echo "オプション:" >&2
    echo "  --title \"タイトル\"  PRのタイトルを更新" >&2
    echo "  --dry-run          実際の更新を行わず、変更内容をプレビュー表示のみ" >&2
    echo "" >&2
    echo "PR本文は標準入力から読み取ります。" >&2
    echo "" >&2
    echo "例:" >&2
    echo "  # 本文のみ更新" >&2
    echo "  $0 123 <<'EOF'" >&2
    echo "  新しいPR本文" >&2
    echo "  EOF" >&2
    echo "" >&2
    echo "  # タイトルと本文を更新" >&2
    echo "  $0 123 --title \"新しいタイトル\" <<'EOF'" >&2
    echo "  新しいPR本文" >&2
    echo "  EOF" >&2
    exit 1
}

# 引数の初期化
PR_NUMBER=""
NEW_TITLE=""
DRY_RUN=false

# 引数解析
while [[ $# -gt 0 ]]; do
    case "$1" in
        --title)
            if [[ $# -lt 2 ]]; then
                echo "エラー: --title オプションには引数が必要です。" >&2
                usage
            fi
            NEW_TITLE="$2"
            shift 2
            ;;
        --dry-run)
            DRY_RUN=true
            shift
            ;;
        -*)
            echo "エラー: 不明なオプション: $1" >&2
            usage
            ;;
        *)
            if [[ -z "$PR_NUMBER" ]]; then
                PR_NUMBER="$1"
            else
                echo "エラー: PR番号は1つだけ指定してください。" >&2
                usage
            fi
            shift
            ;;
    esac
done

# PR番号のチェック
if [[ -z "$PR_NUMBER" ]]; then
    echo "エラー: PR番号を指定してください。" >&2
    usage
fi

# PR番号から # を削除（もしあれば）
PR_NUMBER="${PR_NUMBER#\#}"

# PR番号が数値かチェック
if ! [[ "$PR_NUMBER" =~ ^[0-9]+$ ]]; then
    echo "エラー: PR番号は数値である必要があります: $PR_NUMBER" >&2
    exit 1
fi

# 標準入力から本文を読み取り
if [[ -t 0 ]]; then
    # 標準入力が端末の場合（パイプやリダイレクトがない場合）
    PR_BODY=""
else
    # 標準入力からデータを読み取る
    PR_BODY=$(cat)
fi

# タイトルも本文も指定されていない場合はエラー
if [[ -z "$NEW_TITLE" ]] && [[ -z "$PR_BODY" ]]; then
    echo "エラー: タイトルまたは本文の少なくとも一方を指定してください。" >&2
    echo "  - タイトル更新: --title オプションを使用" >&2
    echo "  - 本文更新: 標準入力から本文を渡す" >&2
    usage
fi

echo "PR #$PR_NUMBER の情報を確認中..." >&2

# PRの存在確認
set +e
PR_INFO=$(gh pr view "$PR_NUMBER" --json number,title,body 2>&1)
EXIT_CODE=$?
set -e

if [ $EXIT_CODE -ne 0 ]; then
    echo "エラー: PR #$PR_NUMBER が見つかりませんでした。" >&2
    echo "$PR_INFO" >&2
    exit 1
fi

# 現在のタイトルと本文を取得
CURRENT_TITLE=$(echo "$PR_INFO" | jq -r '.title')
CURRENT_BODY=$(echo "$PR_INFO" | jq -r '.body // ""')

echo "✓ PR #$PR_NUMBER が見つかりました" >&2
echo "" >&2

# 更新内容の表示
echo "=== 更新内容 ===" >&2

if [[ -n "$NEW_TITLE" ]]; then
    echo "" >&2
    echo "【タイトル】" >&2
    echo "変更前: $CURRENT_TITLE" >&2
    echo "変更後: $NEW_TITLE" >&2
fi

if [[ -n "$PR_BODY" ]]; then
    echo "" >&2
    echo "【本文】" >&2
    echo "変更前:" >&2
    echo "---" >&2
    echo "$CURRENT_BODY" >&2
    echo "---" >&2
    echo "" >&2
    echo "変更後:" >&2
    echo "---" >&2
    echo "$PR_BODY" >&2
    echo "---" >&2
fi

echo "" >&2
echo "================" >&2
echo "" >&2

# ドライランモードの場合は、ここで終了
if [[ "$DRY_RUN" == true ]]; then
    echo "ℹ️  ドライランモード: 実際の更新は行いません。" >&2
    exit 0
fi

# PR更新の実行
echo "PR #$PR_NUMBER を更新中..." >&2

# gh pr edit コマンドの構築
GH_EDIT_CMD=("gh" "pr" "edit" "$PR_NUMBER")

# タイトルが指定されている場合
if [[ -n "$NEW_TITLE" ]]; then
    GH_EDIT_CMD+=("--title" "$NEW_TITLE")
fi

# 本文が指定されている場合
if [[ -n "$PR_BODY" ]]; then
    # 一時ファイルに本文を書き込み
    TEMP_BODY_FILE=$(mktemp)
    trap "rm -f '$TEMP_BODY_FILE'" EXIT
    echo "$PR_BODY" > "$TEMP_BODY_FILE"
    GH_EDIT_CMD+=("--body-file" "$TEMP_BODY_FILE")
fi

# コマンド実行
set +e
UPDATE_RESULT=$("${GH_EDIT_CMD[@]}" 2>&1)
EXIT_CODE=$?
set -e

if [ $EXIT_CODE -ne 0 ]; then
    echo "✗ エラー: PR更新に失敗しました。" >&2
    echo "$UPDATE_RESULT" >&2
    exit 1
fi

echo "✓ PR #$PR_NUMBER を更新しました。" >&2
echo "" >&2
echo "$UPDATE_RESULT" >&2
