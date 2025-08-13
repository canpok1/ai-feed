# ai-feed

AI Feedは、RSSフィードから記事を自動収集し、AIが生成したコメント付きで推薦記事を提供するCLIツールです。

## 🌟 主な機能

- **自動記事収集**: 複数のRSSフィードから最新記事を取得
- **AIコメント生成**: Google Gemini APIを使用して記事に対する洞察に富んだコメントを自動生成
- **多様な出力先**: Slack、Misskey、標準出力への投稿をサポート
- **カスタマイズ可能**: プロンプトテンプレートやメッセージフォーマットを自由に設定
- **設定管理**: YAML形式の設定ファイルとプロファイル機能

## 🎯 ユースケース

- **技術ブログの自動収集**: 複数の技術ブログから最新情報をチェック
- **チーム情報共有**: Slackチャンネルへの定期的な記事共有
- **個人学習**: 興味のある分野の記事をAIコメント付きで効率的に把握
- **SNS投稿**: Misskeyなどの分散SNSへの記事紹介

## こんな人におすすめ

- 技術情報を効率的にキャッチアップしたい開発者
- チームでの情報共有を自動化したい方
- 複数のブログやニュースサイトをフォローしている方
- AIによる記事要約や解説を活用したい方

## 🚀 インストール

### 必要な環境

- Go 1.24以上（ソースからビルドする場合）
- Google Gemini API キー（[取得方法](./docs/05_api_integration.md#google-gemini-api)）

### バイナリのダウンロード（推奨）

#### macOS

```bash
# Intel Mac
wget https://github.com/canpok1/ai-feed/releases/latest/download/ai-feed_darwin_amd64.tar.gz
tar -xzf ai-feed_darwin_amd64.tar.gz
sudo mv ai-feed /usr/local/bin/

# Apple Silicon Mac
wget https://github.com/canpok1/ai-feed/releases/latest/download/ai-feed_darwin_arm64.tar.gz
tar -xzf ai-feed_darwin_arm64.tar.gz
sudo mv ai-feed /usr/local/bin/

# Homebrewでのインストール（将来予定）
# brew install canpok1/tap/ai-feed
```

#### Linux

```bash
# x86_64
wget https://github.com/canpok1/ai-feed/releases/latest/download/ai-feed_linux_amd64.tar.gz
tar -xzf ai-feed_linux_amd64.tar.gz
sudo mv ai-feed /usr/local/bin/

# ARM64
wget https://github.com/canpok1/ai-feed/releases/latest/download/ai-feed_linux_arm64.tar.gz
tar -xzf ai-feed_linux_arm64.tar.gz
sudo mv ai-feed /usr/local/bin/
```

#### Windows

1. [リリース一覧](https://github.com/canpok1/ai-feed/releases)から `ai-feed_windows_amd64.zip` をダウンロード
2. ZIPファイルを解凍
3. `ai-feed.exe` をパスが通ったディレクトリに配置

### Go installを使用したインストール

```bash
go install github.com/canpok1/ai-feed@latest
```

### インストール確認

```bash
ai-feed --version
ai-feed --help
```

## ⚡ クイックスタート

### 1. 設定ファイルの生成

```bash
ai-feed init
```

実行後、カレントディレクトリに `config.yml` が作成されます。

### 2. APIキーの設定

#### 環境変数での設定（推奨）

```bash
# Google Gemini APIキーを設定
export GEMINI_API_KEY="your-api-key-here"

# 永続化する場合（.bashrc, .zshrc等に追記）
echo 'export GEMINI_API_KEY="your-api-key-here"' >> ~/.bashrc
```

#### 設定ファイルでの設定

`config.yml`を編集して、APIキーを設定：

```yaml
default_profile:
  ai:
    gemini:
      type: "gemini-1.5-flash"
      api_key_env: "GEMINI_API_KEY"  # 環境変数名を指定
      # または直接記載（非推奨）
      # api_key: "your-api-key-here"
```

### 3. 初回実行

```bash
# 技術ブログから記事を取得してAIコメント付きで表示
ai-feed recommend --url https://zenn.dev/feed

# 複数のフィードから取得
ai-feed recommend --url https://zenn.dev/feed --url https://qiita.com/popular-items/feed

# ファイルからURL一覧を読み込み
echo "https://zenn.dev/feed" > feeds.txt
echo "https://qiita.com/popular-items/feed" >> feeds.txt
ai-feed recommend --source feeds.txt
```

## ログの色付け

ai-feedはログレベルごとに色を付けて視認性を向上させます：

- **DEBUG**: 灰色
- **INFO**: 緑色  
- **WARN**: 黄色
- **ERROR**: 赤色

### 色の無効化

以下の場合、色付けは自動的に無効化されます：
- パイプやファイルリダイレクト時
- CI環境での実行時
- `NO_COLOR` 環境変数が設定されている時

手動で色を無効化する場合：
```bash
# 色なしでアプリケーションを実行
NO_COLOR=1 ./ai-feed recommend --url {RSS URL}

# 環境変数として設定
export NO_COLOR=1
./ai-feed recommend --url {RSS URL}
```

### サポート環境

ログの色付け機能は以下の環境でサポートされます：
- Unix/Linux
- macOS

Windows環境では色付けされません。

## 📚 詳細な使用方法

### プロファイルの作成

異なる設定を使い分けるためのプロファイル機能：

```bash
# プロファイルファイルを作成
ai-feed profile init my-profile.yml

# プロファイルを使用して実行
ai-feed recommend --profile my-profile.yml --url https://example.com/feed
```

### Slack連携

```bash
# Slack Webhook URLを環境変数に設定
export SLACK_WEBHOOK_URL="https://hooks.slack.com/services/..."

# config.ymlまたはプロファイルに設定を追加
```

```yaml
output:
  slack_api:
    webhook_url_env: "SLACK_WEBHOOK_URL"
    message_template: |
      :newspaper: *{{.Article.Title}}*
      {{.Article.Link}}
      
      {{.Comment}}
```

### Misskey連携

```bash
# Misskey APIトークンを環境変数に設定
export MISSKEY_API_TOKEN="your-token-here"
```

```yaml
output:
  misskey:
    api_url: "https://misskey.io/api"
    api_token_env: "MISSKEY_API_TOKEN"
    message_template: |
      【おすすめ記事】
      {{.Article.Title}}
      {{.Article.Link}}
      
      {{.Comment}}
      
      #ai_feed #tech
```

### よく使うオプション

```bash
# 詳細ログを表示
ai-feed recommend -v --url https://example.com/feed

# 色なしで出力（パイプ時に便利）
NO_COLOR=1 ai-feed recommend --url https://example.com/feed

# 設定チェック
ai-feed profile check config.yml
```

## 🔧 開発者向け情報

本プロジェクトの開発に参加したい方は、以下のドキュメントを参照してください：

- [開発環境セットアップ](./docs/00_development_setup.md) - 開発環境の構築方法
- [アーキテクチャ概要](./docs/02_architecture.md) - プロジェクトの設計と構造
- [テストガイド](./docs/03_testing.md) - テストの書き方と実行方法
- [コントリビューションガイド](./docs/04_contributing.md) - 貢献方法とプルリクエスト作成
- [API連携ガイド](./docs/05_api_integration.md) - 外部API設定の詳細
- [コーディングルール](./docs/01_coding_rules.md) - コーディング規約

### 簡単な開発コマンド

```bash
# プロジェクトをクローンして開発環境をセットアップ
git clone https://github.com/canpok1/ai-feed.git
cd ai-feed
make setup

# テスト実行
make test

# ビルド
make build
```
