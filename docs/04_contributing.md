# コントリビューションガイド

ai-feedプロジェクトへの貢献を歓迎します！このガイドでは、プロジェクトに貢献する方法について説明します。

## 貢献の方法

### バグ報告

1. [Issues](https://github.com/canpok1/ai-feed/issues)で既存の問題を確認
2. 同じ問題が報告されていなければ、新しいIssueを作成
3. バグの詳細、再現手順、期待される動作を記載

### 機能提案

1. [Issues](https://github.com/canpok1/ai-feed/issues)で新しいIssueを作成
2. 提案する機能の概要と必要性を説明
3. 可能であれば実装方法の案も記載

### コードの貢献

プルリクエストを通じてコードを貢献できます。以下の手順に従ってください。

## 開発フロー

### 1. リポジトリのFork

1. GitHubで[ai-feed](https://github.com/canpok1/ai-feed)リポジトリにアクセス
2. 右上の「Fork」ボタンをクリック
3. 自分のアカウントにリポジトリをコピー

### 2. ローカル環境のセットアップ

```bash
# Forkしたリポジトリをクローン
git clone https://github.com/YOUR_USERNAME/ai-feed.git
cd ai-feed

# 上流リポジトリを追加
git remote add upstream https://github.com/canpok1/ai-feed.git

# 開発環境のセットアップ
make setup
```

詳細は[開発環境セットアップガイド](./00_development_setup.md)を参照してください。

### 3. ブランチの作成

```bash
# mainブランチを最新に更新
git checkout main
git pull upstream main

# ブランチを作成
git checkout -b your-branch-name
```

### 4. コードの変更

#### 必須要件

1. **コーディングルールの遵守**
   - [コーディングルール](./01_coding_rules.md)を必ず確認
   - **コメントは全て日本語**で記述

2. **コミット前の必須チェック**
   ```bash
   # コードフォーマット（必須）
   make fmt
   
   # 静的解析（必須）
   make lint
   
   # テストの実行（必須）
   make test
   ```

3. **依存関係の管理**
   ```bash
   # 新しい依存関係を追加した場合
   go mod tidy
   ```

4. **モックの更新**
   ```bash
   # インターフェースを変更した場合
   make generate
   ```

### 5. コミット

```bash
# 変更をコミット
git commit -m "変更内容の説明"
```

### 6. プルリクエストの作成

#### 準備

```bash
# 変更をプッシュ
git push origin your-branch-name

# 上流の変更を取り込む（必要な場合）
git fetch upstream
git rebase upstream/main
```


### 7. コードレビュー

#### レビューを受ける側

- レビューコメントには建設的に対応
- 修正が必要な場合は追加コミットで対応
- 議論が必要な場合はコメントで返信

#### レビューする側

- 建設的なフィードバックを心がける
- コードの動作だけでなく、可読性や保守性も確認
- 良い点も積極的にコメント

## コントリビューター向けの重要事項

### 必ず守るべきルール

1. **コミット前に必ず実行**
   ```bash
   make fmt    # コードフォーマット
   make lint   # 静的解析
   make test   # テスト
   ```

2. **日本語の使用**
   - コメントは全て日本語
   - Issueやプルリクエストの記載も日本語

3. **テストの追加**
   - 新機能には必ずテストを追加
   - バグ修正には再発防止のテストを追加

4. **ドキュメントの更新**
   - APIやコマンドの変更時は必ずドキュメントを更新
   - 新機能はREADMEかdocs/に説明を追加

### 推奨事項

1. **小さなプルリクエスト**
   - 1つのプルリクエストは1つの目的に絞る
   - レビューしやすい適切なサイズに分割

2. **早めのドラフトPR**
   - 実装方針の確認が必要な場合はドラフトPRを作成
   - WIP（Work In Progress）を明記

3. **継続的な更新**
   - mainブランチの変更を定期的に取り込む
   - コンフリクトは早めに解決

## ライセンス

貢献されたコードは、プロジェクトのライセンス（MIT License）に従います。

## 質問・サポート

- 実装方法が不明な場合はIssueで質問
- Discussionsで議論も歓迎
- 日本語でのコミュニケーションを推奨

## 謝辞

ai-feedプロジェクトへの貢献に感謝します！皆様の協力により、より良いツールを提供できます。

## 関連ドキュメント

- [開発環境セットアップ](./00_development_setup.md)
- [コーディングルール](./01_coding_rules.md)
- [アーキテクチャ概要](./02_architecture.md)
- [テストガイド](./03_testing.md)