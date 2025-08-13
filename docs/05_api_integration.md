# 外部API連携ガイド

ai-feedは複数の外部サービスと連携して、AIによる記事推薦とメッセージ配信を実現しています。このドキュメントでは、各APIの設定方法と注意事項を説明します。

## セキュリティに関する重要事項

**APIキーやトークンの取り扱いについて**:
- APIキーを**絶対にコードやGitリポジトリにコミットしない**
- 環境変数または設定ファイルで管理する
- 設定ファイルは`.gitignore`に必ず追加する
- 本番環境では適切な秘密管理システムを使用する
- 定期的なトークンのローテーションを推奨

## Google Gemini API

### 概要
Google Gemini APIは、記事に対するAIコメントを生成するために使用されます。

### APIキーの取得

1. [Google AI Studio](https://makersuite.google.com/app/apikey)にアクセス
2. Googleアカウントでログイン
3. 「Get API Key」をクリック
4. 新しいAPIキーを作成または既存のキーを選択
5. APIキーをコピー

### 設定方法

#### config.ymlでの設定

```yaml
default_profile:
  ai:
    gemini:
      type: "gemini-2.5-flash"  # モデルタイプ
      api_key: "your-api-key-here"  # 直接記載（非推奨）
      # または
      api_key_env: "GEMINI_API_KEY"  # 環境変数名を指定（推奨）
```

#### profile.ymlでの設定

```yaml
ai:
  gemini:
    type: "gemini-2.5-pro"  # より高性能なモデルを使用
    api_key_env: "GEMINI_API_KEY"

# 出力先の個別制御
output:
  slack_api:
    enabled: true  # Slackへの投稿を有効化
    api_token_env: "SLACK_API_TOKEN"
    channel: "#tech-news"
  
  misskey:
    enabled: false  # Misskeyへの投稿を無効化（APIトークン不要）
```

#### 環境変数での設定（推奨）

```bash
# シェルで設定
export GEMINI_API_KEY="your-actual-api-key"
```

### 利用可能なモデル

| モデル名 | 特徴 | レート制限 |
|---------|------|-----------|
| gemini-2.5-flash | 高速・低コスト | 15 RPM (無料版) |
| gemini-2.5-flash-8b | より高速 | 15 RPM (無料版) |
| gemini-2.5-pro | 高精度 | 2 RPM (無料版) |

### プロンプトのカスタマイズ

```yaml
prompt:
  system_prompt: |
    あなたは技術ブログの専門家です。
    記事を読んで、開発者向けの有益なコメントを生成してください。
  
  comment_prompt_template: |
    以下の記事について、技術的な観点から200文字程度でコメントしてください。
    
    タイトル: {{TITLE}}
    URL: {{URL}}
    内容: {{CONTENT}}
```

### エラーハンドリング

Gemini APIのエラーコード：
- `429`: レート制限超過 → 待機後リトライ
- `401`: 認証エラー → APIキーを確認
- `500`: サーバーエラー → 自動リトライ

## Slack API

### 概要
Slack APIを使用して、推薦記事をSlackチャンネルに投稿します。Bot TokenとWeb APIを使用してメッセージを送信します。

### APIトークンの取得

1. [Slack Apps](https://api.slack.com/apps)にアクセス
2. 「Create New App」をクリック
3. 「From scratch」を選択
4. アプリ名とワークスペースを選択
5. 「OAuth & Permissions」セクションに移動
6. Bot Token Scopesで以下の権限を追加：
   - `chat:write` - メッセージの投稿（必須）
   - `chat:write.public` - 参加していないチャンネルへの投稿
7. 「Install to Workspace」をクリック
8. Bot User OAuth Tokenをコピー（xoxb-で始まるトークン）

### 設定方法

#### config.ymlでの設定

```yaml
default_profile:
  output:
    slack_api:
      enabled: true  # 有効/無効フラグ（省略時はtrue）
      api_token: "xoxb-your-bot-token-here"  # 直接記載（非推奨）
      # または
      api_token_env: "SLACK_API_TOKEN"  # 環境変数名を指定（推奨）
      channel: "#general"  # 投稿先チャンネル
      message_template: |
        {{.Comment}}
        <{{.Article.Link}}|{{.Article.Title}}>
        {{.FixedMessage}}
```

### メッセージテンプレート

利用可能な変数：
- `{{.Article.Title}}`: 記事タイトル
- `{{.Article.Link}}`: 記事URL
- `{{.Comment}}`: AIによるコメント
- `{{.FixedMessage}}`: 固定メッセージ

テンプレートエイリアス（簡単記法）：
- `{{TITLE}}`: `{{.Article.Title}}`の短縮形
- `{{URL}}`: `{{.Article.Link}}`の短縮形
- `{{COMMENT}}`: `{{.Comment}}`の短縮形
- `{{FIXED_MESSAGE}}`: `{{.FixedMessage}}`の短縮形

### Slackマークアップ

```yaml
message_template: |
  *{{TITLE}}*  # 太字
  _{{COMMENT}}_  # イタリック
  `code`  # コード
  >引用  # 引用
  :emoji:  # 絵文字
  <{{URL}}|{{TITLE}}>  # リンク付きテキスト
```


## Misskey API

### 概要
Misskey APIを使用して、推薦記事をMisskeyインスタンスに投稿します。

### APIトークンの取得

1. Misskeyインスタンスにログイン
2. 設定 → API → アクセストークンの発行
3. 必要な権限を選択：
   - `write:notes` - ノートの投稿（必須）
4. トークンを生成してコピー

### 設定方法

#### config.ymlでの設定

```yaml
default_profile:
  output:
    misskey:
      api_url: "https://misskey.io"  # インスタンスのAPI URL
      api_token: "your-token-here"  # 直接記載（非推奨）
      # または
      api_token_env: "MISSKEY_API_TOKEN"  # 環境変数名を指定（推奨）
      message_template: |
        【おすすめ記事】
        {{.Article.Title}}
        {{.Article.Link}}
        
        {{.Comment}}
        
        #ai_feed #tech
```

## 複数の出力先設定

### 同時投稿

```yaml
output:
  slack_api:
    enabled: true
    api_token_env: "SLACK_API_TOKEN"
    channel: "#general"
    message_template: "{{COMMENT}}\n<{{URL}}|{{TITLE}}>"
  
  misskey:
    enabled: true
    api_url: "https://misskey.io"
    api_token_env: "MISSKEY_API_TOKEN"
    message_template: "{{COMMENT}}\n{{TITLE}}\n{{URL}}"
```

両方の出力先に同時に投稿されます。

### 条件付き投稿

プロファイルを使い分けることで、状況に応じた投稿先の切り替えが可能：

```bash
# Slackのみに投稿
./ai-feed recommend --profile slack-only.yml --url https://example.com/feed

# Misskeyのみに投稿
./ai-feed recommend --profile misskey-only.yml --url https://example.com/feed
```

## エラーハンドリングとリトライ

### エラー時の動作

- 1つの出力先でエラーが発生しても、他の出力先への投稿は継続
- 全てのエラーはログに記録
- 致命的なエラー（認証失敗など）は即座に停止

## デバッグとトラブルシューティング

### verbose モードでの実行

```bash
# 詳細なログを表示
./ai-feed recommend -v --url https://example.com/feed
```

## ベストプラクティス

### 1. 環境変数の使用

```bash
# 環境変数の設定例
export GEMINI_API_KEY="your-gemini-api-key"
export SLACK_API_TOKEN="xoxb-your-slack-bot-token"
export MISSKEY_API_TOKEN="your-misskey-token"

# または、起動時に環境変数を指定
GEMINI_API_KEY="your-key" ./ai-feed recommend --url https://example.com/feed
```

**注意**: ai-feedは.envファイルの自動読み込みには対応していません。環境変数は手動で設定する必要があります。

### 2. プロファイルの分離

```bash
# 本番用
profiles/production.yml

# 開発用
profiles/development.yml

# テスト用
profiles/test.yml
```

### 3. セキュアな運用

- APIキーは最小権限の原則に従う
- 定期的なトークンローテーション
- アクセスログの監視
- 異常なアクセスパターンの検知

### 4. エラー監視

```yaml
# ログ出力先の設定（将来実装予定）
logging:
  level: "info"
  output: "/var/log/ai-feed/app.log"
  error_output: "/var/log/ai-feed/error.log"
```

## 関連ドキュメント

- [アーキテクチャ概要](./02_architecture.md) - API連携の内部実装
- [開発環境セットアップ](./00_development_setup.md) - 開発環境での設定
- [テストガイド](./03_testing.md) - API連携のテスト方法