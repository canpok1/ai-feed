# スクリプトガイド

このドキュメントでは、`scripts/` ディレクトリに配置されている便利なスクリプトについて説明します。

## 概要

プロジェクトには、開発・運用作業を効率化するための各種シェルスクリプトが用意されています。これらのスクリプトは、バージョン管理、プルリクエストのレビュー対応などをサポートします。

## 利用可能なスクリプト

### create-version-tag.sh

mainブランチにバージョンタグを付与するスクリプトです。

#### 機能
- 最新のリリースバージョンのパッチバージョンを1つ進めたタグを自動作成
- セマンティックバージョニング（v1.2.3形式）に準拠
- ドライランモードで事前確認が可能

#### 使用方法

```bash
# 実際にタグを作成してプッシュ
./scripts/create-version-tag.sh

# ドライランモード（タグ作成・プッシュをスキップ）
./scripts/create-version-tag.sh --dry-run
```

#### 動作例

```bash
$ ./scripts/create-version-tag.sh --dry-run
最新のタグ: v0.1.5
新しいバージョン: v0.1.6
[ドライラン] タグの作成とプッシュをスキップします。

$ ./scripts/create-version-tag.sh
最新のタグ: v0.1.5
新しいバージョン: v0.1.6
タグ v0.1.6 を作成しました。
タグ v0.1.6 をプッシュしました。
```

#### 注意事項
- mainブランチでの実行を推奨
- 初回実行時（タグが存在しない場合）は v0.0.1 を作成
- プレリリースタグ（例: v1.0.0-alpha）は除外してカウント

### get-pr-review-comments.sh

プルリクエストの未解決レビューコメントを取得するスクリプトです。

#### 機能
- 指定したPRの未解決レビュースレッドを一覧表示
- スレッドID、コメント作成者、コメント内容をJSON形式で出力
- GitHub GraphQL APIを使用

#### 前提条件
- GitHub CLI（`gh`）がインストール済みで認証済みであること
- `jq`コマンドがインストール済みであること

#### 使用方法

```bash
# PR番号を指定して未解決コメントを取得
./scripts/get-pr-review-comments.sh <PR番号>

# 例
./scripts/get-pr-review-comments.sh 123
```

#### 出力例

現在の実装では、複数のJSONオブジェクトが改行区切りで出力されます（NDJSON形式）:

```json
{
  "thread_id": "PRRT_kwDONTZR484BhKaH",
  "author": "reviewer-username",
  "comment": "この部分のエラーハンドリングを改善してください。"
}
{
  "thread_id": "PRRT_kwDONTZR484BhKaI",
  "author": "another-reviewer",
  "comment": "テストケースを追加してください。"
}
```

**注**: より標準的なJSON配列形式への出力に変更することも検討できます。その場合の出力イメージ:

```json
[
  {
    "thread_id": "PRRT_kwDONTZR484BhKaH",
    "author": "reviewer-username",
    "comment": "この部分のエラーハンドリングを改善してください。"
  },
  {
    "thread_id": "PRRT_kwDONTZR484BhKaI",
    "author": "another-reviewer",
    "comment": "テストケースを追加してください。"
  }
]
```

#### 注意事項
- 最大30件のレビュースレッドを取得
- 各スレッドの最初の10件のコメントを取得
- PR番号は数値で指定

### resolve-review-thread.sh

レビュースレッドを解決済み（resolved）にするスクリプトです。

#### 機能
- 指定したスレッドIDのレビュースレッドをresolvedに変更
- GitHub GraphQL APIを使用してスレッドを更新

#### 前提条件
- GitHub CLI（`gh`）がインストール済みで認証済みであること
- `jq`コマンドがインストール済みであること
- スレッドをresolveする権限があること

#### 使用方法

```bash
# スレッドIDを指定してresolve
./scripts/resolve-review-thread.sh <スレッドID>

# 例
./scripts/resolve-review-thread.sh "PRRT_kwDONTZR484BhKaH"
```

#### 出力例

```bash
レビュースレッドをresolve中...
スレッドID: PRRT_kwDONTZR484BhKaH

✓ レビュースレッドをresolveしました。
スレッドID: PRRT_kwDONTZR484BhKaH
```

#### 注意事項
- スレッドIDは `get-pr-review-comments.sh` の出力から取得可能
- スレッドIDは引用符で囲んで指定すること
- 既にresolve済みのスレッドに対しても実行可能

### reply-to-review-thread.sh

レビュースレッドに返信を投稿するスクリプトです。

#### 機能
- 指定したスレッドIDのレビュースレッドに返信を投稿
- コメント投稿者への@メンションを自動的に追加
- 標準入力からコメント本文を読み取り
- GitHub GraphQL APIを使用

#### 前提条件
- GitHub CLI（`gh`）がインストール済みで認証済みであること
- `jq`コマンドがインストール済みであること
- スレッドにコメントを投稿する権限があること

#### 使用方法

```bash
# パイプ経由でコメントを投稿
echo "コメント内容" | ./scripts/reply-to-review-thread.sh <スレッドID>

# ヒアドキュメントで複数行コメントを投稿
./scripts/reply-to-review-thread.sh <スレッドID> <<EOF
複数行の
コメント内容
EOF

# 例
echo "ご指摘ありがとうございます。修正しました。" | ./scripts/reply-to-review-thread.sh "PRRT_kwDONTZR484BhKaH"
```

#### 出力例

```bash
コメント本文を読み取り中...
リポジトリ情報を取得中...
リポジトリ: owner/ai-feed
スレッドID: PRRT_kwDONTZR484BhKaH

スレッド情報を取得中...
コメント投稿者: @reviewer-username

返信を投稿中...

✓ 返信を投稿しました。
コメントID: PRRC_kwDONTZR484BhKaJ
投稿者: @your-username
作成日時: 2025-12-13T10:30:00Z
```

#### 注意事項
- スレッドIDは `get-pr-review-comments.sh` の出力から取得可能
- コメント投稿者への@メンションは自動的に追加されるため、手動で記述する必要はありません
- 標準入力からコメントを読み取るため、空のコメントは投稿できません
- スレッドIDはGitHub GraphQL APIのNode ID形式（例: `PRRT_kwDO...`）で指定

## 典型的なワークフロー

### プルリクエストレビュー対応

1. **未解決コメントの確認**
   ```bash
   ./scripts/get-pr-review-comments.sh 123
   ```

2. **指摘事項の修正**
   - コードを修正してコミット・プッシュ

3. **レビュースレッドへの返信**（オプション）
   ```bash
   echo "ご指摘ありがとうございます。修正しました。" | ./scripts/reply-to-review-thread.sh "PRRT_kwDONTZR484BhKaH"
   ```

4. **スレッドの解決**
   ```bash
   ./scripts/resolve-review-thread.sh "PRRT_kwDONTZR484BhKaH"
   ```

### リリースバージョンタグの作成

1. **ドライランで確認**
   ```bash
   ./scripts/create-version-tag.sh --dry-run
   ```

2. **タグの作成とプッシュ**
   ```bash
   ./scripts/create-version-tag.sh
   ```

## トラブルシューティング

### GitHub CLI認証エラー

```bash
# GitHub CLIの認証状態を確認
gh auth status

# 再認証が必要な場合
gh auth login
```

### jqコマンドが見つからない

```bash
# macOS
brew install jq

# Linux (Ubuntu/Debian)
sudo apt-get install jq

# Linux (CentOS/RHEL/Fedora)
sudo yum install jq
```

### PR番号が見つからない

- PR番号が正しいか確認
- プライベートリポジトリの場合、適切なアクセス権限があるか確認
- GitHub CLIが正しいリポジトリを参照しているか確認
  ```bash
  gh repo view
  ```

## 関連ドキュメント

- [開発環境セットアップガイド](./00_development_setup.md) - GitHub CLIのセットアップ方法
- [コントリビューションガイド](./04_contributing.md) - プルリクエストの作成とレビュー対応

