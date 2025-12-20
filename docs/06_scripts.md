# スクリプトガイド

このドキュメントでは、`scripts/` ディレクトリに配置されている便利なスクリプトについて説明します。

## 概要

プロジェクトには、開発・運用作業を効率化するための各種シェルスクリプトが用意されています。これらのスクリプトは、バージョン管理などをサポートします。

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

## 典型的なワークフロー

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

## 関連ドキュメント

- [開発環境セットアップガイド](./00_development_setup.md) - GitHub CLIのセットアップ方法
- [コントリビューションガイド](./04_contributing.md) - プルリクエストの作成とレビュー対応
