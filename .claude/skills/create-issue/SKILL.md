---
name: create-issue
description: GitHub issueを作成します。タイトルと本文を受け取り、gh issue createコマンドでissueを作成します。
---

## 使用方法

### 作業手順

1. **issue説明文を準備する（推奨）**
   - document-specialistエージェントでissue説明文を整理することを推奨
   - 問題、詳細、期待される結果を明確に記述

2. **リポジトリ情報を確認**
   - `git remote -v` でリポジトリ情報を確認
   - 現在のリポジトリでissueを作成することを確認

3. **gh issue createコマンドを実行**
   - タイトルと本文を引数として受け取る
   - `gh issue create --title "タイトル" --body "本文"`を実行
   - 作成されたissue URLを表示

### 引数

- **タイトル**: issueのタイトル（必須）
- **本文**: issueの詳細説明（必須）

## 典型的なワークフロー

1. **issue説明文の整理（推奨）**
   - document-specialistエージェントを使用
   - 問題、詳細、期待される結果を含む説明文を作成

2. **リポジトリ確認**
   ```bash
   git remote -v
   ```

3. **issue作成**
   ```bash
   gh issue create --title "タイトル" --body "本文"
   ```

4. **作成確認**
   - 表示されたissue URLで内容を確認

## 実行例

```bash
# Claude Codeがスキル実行時に行う操作
BODY=$(cat <<<'EOS'
## 問題・要望

現在、ユーザー認証機能が実装されていません。

## 詳細

- OAuth 2.0による認証を実装
- GitHub、Google アカウントでのログインをサポート

## 期待される結果

ユーザーが外部アカウントでログインできる
EOS
)
gh issue create \
  --title "新機能: ユーザー認証機能の追加" \
  --body "$BODY"
```

## 注意事項

- タイトルと本文は必須です
- document-specialistエージェントで説明文を事前に整理することを推奨します
- issue作成後、URLが表示されるので内容を確認してください
- リポジトリの権限がない場合はエラーになります

## 関連ドキュメント

- [文書化ルール](../../../docs/07_document_rules.md)
- [コントリビューションガイド](../../../docs/04_contributing.md)
