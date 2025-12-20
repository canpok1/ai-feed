---
name: plan-issue
description: GitHub issue情報を取得し、対応用のタスクファイル作成をサポートします。スクリプトがissue情報を取得し、Claude Codeがその情報を元にtmp/todoフォルダにタスクファイルを作成します。
---

## 使用方法

### スクリプトの実行

```bash
./.claude/skills/plan-issue/plan-issue.sh <ISSUE_NUMBER>
```

### 引数

- `<ISSUE_NUMBER>`: GitHub issueの番号（必須、#付きでも無しでも可）
  - 例: `"#123"` または `"123"`

### 実行例

```bash
# issue #123の情報を取得
./.claude/skills/plan-issue/plan-issue.sh "#123"

# または
./.claude/skills/plan-issue/plan-issue.sh "123"
```

## タスクファイル作成ガイド

スクリプトがissue情報を取得した後、Claude Codeは以下のガイドラインに従って`tmp/todo`フォルダにタスクファイルを作成します。

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
   ```bash
   ./.claude/skills/plan-issue/plan-issue.sh "#123"
   ```

2. **タスクファイルの作成**
   - Claude Codeが取得したissue情報を分析
   - tmp/todoフォルダに適切な数のタスクファイルを作成
   - 各タスクファイルに具体的な作業内容を記載

3. **タスクの実行**
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
