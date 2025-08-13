---
argument-hint: [issue番号（#付き)]
description: github issue 対応用の開発タスクを作成する
---

github issue $ARGUMENTS の対応を行うための作業手順を考えて、
1ファイル1タスクとして tmp/todo フォルダにファイルを作成してください。
リポジトリ情報は `git remote -v` で確認すること。

## タスクファイルのフォーマット

### ファイル名

issue_{github issueの番号}_plan_{2桁0埋めの1からの連番}_{タスク概要(英語)}.md

例）issue #1 に対するタスクの場合、 `issue_1_plan_01_sample_task.md` となる

### ファイル内容のテンプレート

```
## 対応内容の概要

## 対応内容の詳細

### 編集対象ファイル

### 完了条件

### 備考
- 適当な粒度でコミットすること。
```
