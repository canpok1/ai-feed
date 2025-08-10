---
argument-hint: [PR番号（#付き)]
description: プルリクエスト対応用の開発タスクを作成する
---

PR $ARGUMENTS にレビューコメントが投稿されました。
コメントを確認して対応要否を判断し、対応が必要なものを1ファイル1タスクとして tmp/todo フォルダにファイルを作成してください。

## コメントの確認手順

```
OWNER_REPO=$(gh repo view --json nameWithOwner --jq '.nameWithOwner')
OWNER=$(echo $OWNER_REPO | cut -d'/' -f1)
REPO=$(echo $OWNER_REPO | cut -d'/' -f2)
PR_NUMBER=$(gh pr view --json number --jq '.number')

gh api graphql -f query="
{
  repository(owner: \"${OWNER}\", name: \"${REPO}\") {
    pullRequest(number: ${PR_NUMBER}) {
      reviewThreads(first: 30) {
        nodes {
          id
          isResolved
          comments(first: 10) {
            nodes {
              id
              body
              author {
                login
              }
            }
          }
        }
      }
    }
  }
}" --jq '.data.repository.pullRequest.reviewThreads.nodes[] | select(.isResolved == false) | {thread_id: .id, author: 
.comments.nodes[0].author.login, comment: .comments.nodes[0].body}'
```

## タスクファイルのフォーマット

### ファイル名

pr_{PRの番号}_task_{2桁0埋めの1からの連番}_{タスク概要(英語)}.md

例）PR #1 に対するタスクの場合、 `pr_1_task_01_sample_task.md` となる

### ファイル内容のテンプレート

```
## 対応内容の概要

## 対応内容の詳細

### レビューコメント情報

- PR番号: {PRの番号を記載する}
- レビュースレッドID: {レビュースレッドIDを記載する}
- 投稿者: {レビューコメントの投稿者名を記載する}

### 編集対象ファイル

### 完了条件

### 備考
- レビューコメント投稿者がgemini-code-assistの場合、対応完了後にコミットとpushを行いレビューコメントをresolveすること。
    - resolveするためのコマンド。THREAD_IDはレビュースレッドIDに置き換えること。
        - `gh api graphql -f query='mutation { resolveReviewThread(input: {threadId: "THREAD_ID"}) { thread { id isResolved } } }'`
```
