---
argument-hint: [issue番号（#付き)]
description: github issue の内容を精緻化する
---

github issue $ARGUMENTS の内容を確認し、以下を洗い出してください。
- 仕様の表現が曖昧な点
- 実装上の仕様の不明点

その後、 github issue の内容を更新するための作業タスクを1ファイル1タスクとして tmp/todo フォルダにファイルを作成してください。
リポジトリ情報は `git remote -v` で確認すること。

## タスクファイルのフォーマット

### ファイル名

issue_{github issueの番号}_refine_{2桁0埋めの1からの連番}_{タスク概要(英語)}.md

例）issue #1 に対するタスクの場合、 `issue_1_refine_01_sample_task.md` となる

### ファイル内容のテンプレート

```
## 対応内容の概要

## 対応内容の詳細

### 編集対象のgithub issueのURL

### 完了条件

### 備考
- 本タスクではgithub issue以外（ソースコードなど）の編集は禁止
- github issue の更新は `gh issue edit` コマンドで行うこと
```
