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
curl -sLO https://github.com/canpok1/ai-feed/releases/latest/download/ai-feed_darwin_amd64.tar.gz
tar -xzf ai-feed_darwin_amd64.tar.gz
sudo mv ai-feed /usr/local/bin/
rm ai-feed_darwin_amd64.tar.gz

# Apple Silicon Mac
curl -sLO https://github.com/canpok1/ai-feed/releases/latest/download/ai-feed_darwin_arm64.tar.gz
tar -xzf ai-feed_darwin_arm64.tar.gz
sudo mv ai-feed /usr/local/bin/
rm ai-feed_darwin_arm64.tar.gz

# Homebrewでのインストール（将来予定）
# brew install canpok1/tap/ai-feed
```

#### Linux

```bash
# x86_64
wget https://github.com/canpok1/ai-feed/releases/latest/download/ai-feed_linux_amd64.tar.gz
tar -xzf ai-feed_linux_amd64.tar.gz
sudo mv ai-feed /usr/local/bin/
rm ai-feed_linux_amd64.tar.gz

# ARM64
wget https://github.com/canpok1/ai-feed/releases/latest/download/ai-feed_linux_arm64.tar.gz
tar -xzf ai-feed_linux_arm64.tar.gz
sudo mv ai-feed /usr/local/bin/
rm ai-feed_linux_arm64.tar.gz
```

#### Windows

1. [リリース一覧](https://github.com/canpok1/ai-feed/releases)から `ai-feed_windows_amd64.zip` をダウンロード
2. ZIPファイルを解凍
3. `ai-feed.exe` をパスが通ったディレクトリに配置

### Go installを使用したインストール

```bash
go install github.com/canpok1/ai-feed@latest
```

**注意**: `go install` でインストールされたバイナリは `$GOPATH/bin` または `$HOME/go/bin` に配置されます。このディレクトリが PATH に含まれていない場合は、以下のコマンドで PATH を設定してください：

```bash
# 現在のセッションのみ有効
export PATH=$PATH:$(go env GOPATH)/bin

# 永続化する場合（.bashrc, .zshrc等に追記）
echo 'export PATH=$PATH:$(go env GOPATH)/bin' >> ~/.bashrc
```

### インストール確認

```bash
ai-feed --help
```

### 更新方法

最新版へ更新する場合は、以下の手順で行ってください：

1. [GitHub Releases](https://github.com/canpok1/ai-feed/releases)にアクセス
2. 最新版のバイナリをダウンロード
3. 既存のバイナリ（例: `/usr/local/bin/ai-feed`）を新しいバイナリで置き換え

または、以下のコマンドでインストールし直すことも可能です：
```bash
go install github.com/canpok1/ai-feed@latest
```

## 🛠️ 利用可能なコマンド

### メインコマンド

| コマンド | 説明 |
|----------|------|
| `ai-feed init` | 設定ファイル（config.yml）を生成 |
| `ai-feed recommend` | 記事の推薦とコメント投稿を実行 |
| `ai-feed version` | バージョン情報を表示 |

### 設定管理コマンド

| コマンド | 説明 |
|----------|------|
| `ai-feed config check` | 設定ファイル（config.yml）を検証 |
| `ai-feed config check -v` | 設定ファイルを検証（詳細サマリー表示） |
| `ai-feed profile init <file>` | 新しいプロファイルファイルを作成 |
| `ai-feed profile check [file]` | プロファイルファイルを検証 |

詳細なオプションについては `ai-feed [コマンド] --help` でご確認ください。

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
      type: "gemini-2.5-flash"
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

### 設定リファレンス

設定値の必須/任意について理解しておくと、設定ファイルを正しく作成できます。

#### 設定値の分類

| 表記 | 説明 |
|------|------|
| **必須** | 省略するとバリデーションエラーになります |
| **任意** | 省略可能です。デフォルト値が使用されるか、機能が無効になります |
| **条件付き必須** | 特定の条件下で必須となります |

#### 設定値一覧

| 設定項目 | 必須/任意 | デフォルト値 | 説明 |
|----------|----------|--------------|------|
| `ai.gemini.type` | 必須 | - | 使用するGeminiモデル名 |
| `ai.gemini.api_key` または `api_key_env` | 必須（どちらか） | - | Gemini APIキー |
| `ai.mock.enabled` | 任意 | `false` | モックAIの有効/無効（テスト用） |
| `ai.mock.selector_mode` | 任意 | `first` | 記事選択モード（`first`, `random`, `last`） |
| `ai.mock.comment` | 任意 | 空文字列 | モックが返す固定コメント |
| `system_prompt` | 必須 | - | AIの性格を定義するプロンプト |
| `comment_prompt_template` | 必須 | - | 記事紹介文生成用テンプレート |
| `selector_prompt` | 必須 | - | 記事選択用プロンプト |
| `fixed_message` | 任意 | 空文字列 | メッセージに追加する固定文言 |
| `output.slack_api.enabled` | 任意 | `true` | Slack投稿の有効/無効 |
| `output.slack_api.api_token`/`api_token_env` | 条件付き必須 | - | enabled=trueの場合必須 |
| `output.slack_api.channel` | 条件付き必須 | - | enabled=trueの場合必須 |
| `output.slack_api.message_template` | 条件付き必須 | - | enabled=trueの場合必須 |
| `output.slack_api.username` | 任意 | - | Bot表示名 |
| `output.slack_api.icon_url` | 任意 | - | アイコンURL（icon_emojiと併用不可） |
| `output.slack_api.icon_emoji` | 任意 | - | アイコン絵文字（icon_urlと併用不可） |
| `output.misskey.enabled` | 任意 | `true` | Misskey投稿の有効/無効 |
| `output.misskey.api_token`/`api_token_env` | 条件付き必須 | - | enabled=trueの場合必須 |
| `output.misskey.api_url` | 条件付き必須 | - | enabled=trueの場合必須 |
| `output.misskey.message_template` | 条件付き必須 | - | enabled=trueの場合必須 |
| `cache.enabled` | 任意 | `false` | キャッシュ機能の有効/無効 |
| `cache.file_path` | 任意 | `~/.ai-feed/recommend_history.jsonl` | キャッシュファイルのパス |
| `cache.max_entries` | 任意 | `1000` | 最大エントリ数 |
| `cache.retention_days` | 任意 | `30` | 保持期間（日数） |

#### APIキー・トークン設定について

APIキーやトークンは、直接指定（`api_key`/`api_token`）と環境変数指定（`api_key_env`/`api_token_env`）の2通りの方法で設定できます。

| 設定方法 | 説明 |
|----------|------|
| `api_key` / `api_token` | 設定ファイルに直接値を記載（非推奨） |
| `api_key_env` / `api_token_env` | 環境変数名を指定し、実行時に環境変数から値を読み込む（推奨） |

**動作ルール:**
- 両方が指定された場合、直接指定（`api_key`/`api_token`）が優先されます
- `api_key_env`/`api_token_env`で指定した環境変数が未設定の場合、エラーになります

#### profile checkコマンドの検証ルール

`profile check [file]` コマンドは以下の順序で検証を行います:

1. **config.ymlの読み込み**: デフォルトプロファイル（`default_profile`）を読み込みます
2. **プロファイルのマージ**: プロファイルファイルが指定されている場合、その設定がデフォルトプロファイルを上書きします
3. **バリデーション実行**: マージされた最終的なプロファイルに対して検証を実行します

**重要**: プロファイルファイル単体で全ての必須項目を満たす必要はありません。config.ymlのデフォルトプロファイルと合わせて必須項目が揃っていればOKです。

```
# 例: config.ymlでAI設定を定義し、プロファイルでoutput設定だけを変更する場合

# config.yml
default_profile:
  ai:
    gemini:
      type: gemini-2.5-flash
      api_key_env: GEMINI_API_KEY
  system_prompt: "..."
  comment_prompt_template: "..."
  selector_prompt: "..."
  output:
    slack_api:
      enabled: false
    misskey:
      enabled: false

# my-profile.yml（AI設定は省略可能）
output:
  slack_api:
    enabled: true
    api_token_env: SLACK_TOKEN
    channel: "#tech-news"
    message_template: "..."
```

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
# Slack API トークンを環境変数に設定
export SLACK_TOKEN="your-token-here"

# config.ymlまたはプロファイルに設定を追加
```

```yaml
output:
  slack_api:
    api_token_env: "SLACK_TOKEN"
    channel: "#general"
    message_template: |
      :newspaper: *{{.Article.Title}}*
      {{.Article.Link}}
      
      {{.Comment}}
```

### Misskey連携

```bash
# Misskey APIトークンを環境変数に設定
export MISSKEY_TOKEN="your-token-here"
```

```yaml
output:
  misskey:
    api_url: "https://misskey.io"
    api_token_env: "MISSKEY_TOKEN"
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

# 設定ファイルの検証
ai-feed config check

# 設定ファイルの詳細検証（サマリー表示付き）
ai-feed config check -v

# プロファイルファイルの検証
ai-feed profile check my-profile.yml
```

## 🔧 開発者向け情報

本プロジェクトの開発に参加したい方は、以下のドキュメントを参照してください：

- [開発環境セットアップ](./docs/00_development_setup.md) - 開発環境の構築方法
- [アーキテクチャ概要](./docs/02_architecture_rules.md) - プロジェクトの設計と構造
- [テストガイド](./docs/03_testing_rules.md) - テストの書き方と実行方法
- [コントリビューションガイド](./docs/04_contributing.md) - 貢献方法とプルリクエスト作成
- [API連携ガイド](./docs/05_api_integration.md) - 外部API設定の詳細
- [コーディングルール](./docs/01_coding_rules.md) - コーディング規約
- [その他](https://ktnet.info/ai-feed/) - カバレッジレポートなど

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

### その他の便利なコマンド

```bash
# 開発中にアプリケーションを実行
# option="..." には ai-feed コマンドの引数を渡します
make run option="recommend"
make run option="recommend --url https://example.com/feed"

# リリースビルドをローカルでテスト（goreleaserによるクロスコンパイル）
make build-release

# ビルド成果物とテストファイルを削除
make clean

# インターフェース変更後にモックを再生成
make generate
```
