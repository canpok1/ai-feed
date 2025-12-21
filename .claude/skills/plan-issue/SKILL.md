---
name: plan-issue
description: GitHub issue情報を取得し、対応用のタスクファイル作成をサポートします。スクリプトがissue情報を取得し、Claude Codeがその情報を元にtmp/todoフォルダにタスクファイルを作成します。
---

## 使用方法

### issue情報の取得

Claude Codeは以下のGitHub CLIコマンドを使用してissue情報を取得します：

```bash
gh issue view <ISSUE_NUMBER> --json number,title,body,state,labels,assignees,milestone,createdAt,updatedAt
```

### 引数

- `<ISSUE_NUMBER>`: GitHub issueの番号（必須、#付きでも無しでも可）
  - 例: `"#123"` または `"123"`

### 取得される情報

JSON出力には以下のフィールドが含まれます：

| フィールド | 説明 |
|-----------|------|
| `number` | issue番号 |
| `title` | タイトル |
| `body` | 本文 |
| `state` | 状態（OPEN/CLOSED） |
| `labels` | ラベルオブジェクトの配列。各オブジェクトの`name`フィールドからラベル名を取得します。 |
| `assignees` | 担当者オブジェクトの配列。各オブジェクトの`login`フィールドから担当者名を取得します。 |
| `milestone` | マイルストーンオブジェクト。`title`フィールドからマイルストーン名を取得します。 |
| `createdAt` | 作成日時（ISO 8601形式） |
| `updatedAt` | 更新日時（ISO 8601形式） |

## タスクファイル作成ガイド

issue情報を取得した後、Claude Codeは以下のガイドラインに従って`tmp/todo`フォルダにタスクファイルを作成します。

### 作業方針

- **テスト駆動開発（TDD）**で作業を行うこと
- **1ファイル1タスク**として分割すること
- リポジトリ情報は `git remote -v` で確認すること

### ファイル名フォーマット

```
issue_{GitHub issueの番号}_plan_{2桁0埋めの1からの連番}_{タスク概要(英語)}.md
```

**例**: issue #1に対するタスクの場合

```
issue_1_plan_01_create_test.md
issue_1_plan_02_implement_feature.md
issue_1_plan_03_update_docs.md
```

### ファイル内容のテンプレート

```markdown
## 対応内容の概要

## 対応内容の詳細

### 編集対象ファイル

### 完了条件

### 備考
- 適当な粒度でコミットすること。
```

## 典型的なワークフロー

1. **issue情報の取得**
   - Claude Codeがgh issue viewコマンドを実行
   - JSON出力を解析して必要な情報を抽出

2. **作業計画の策定**
   - `planner` エージェントを使用して作業計画を策定
   - issue情報をplannerエージェントに渡し、並列化可能なタスク分解を依頼

3. **タスクファイルの作成**
   - plannerエージェントの出力に基づいてタスクファイルを作成
   - tmp/todoフォルダに適切な数のタスクファイルを作成
   - 各タスクファイルに具体的な作業内容を記載

4. **タスクの実行**
   - 作成されたタスクファイルに従って実装を進める
   - TDDに従い、テストを先に書いてから実装
   - 適切な粒度でコミット

## 注意事項

- tmp/todoフォルダが存在しない場合は自動的に作成されます
- issue情報の取得には GitHub CLI (`gh`) が必要です
- issueが存在しない場合はエラーになります
- タスクの粒度は、issue の内容に応じて適切に調整してください

## 関連ドキュメント

- [コントリビューションガイド](../../../docs/04_contributing.md) - プルリクエストの作成方法
- [テストルール](../../../docs/03_testing_rules.md) - TDDの実践方法
