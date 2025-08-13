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
```

#### 環境変数での設定（推奨）

```bash
# .envファイルを作成
echo "GEMINI_API_KEY=your-actual-api-key" >> .env

# またはシェルで設定
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
    
    タイトル: {{.Title}}
    概要: {{.Description}}
    URL: {{.Link}}
```

### エラーハンドリング

Gemini APIのエラーコード：
- `429`: レート制限超過 → 待機後リトライ
- `401`: 認証エラー → APIキーを確認
- `500`: サーバーエラー → 自動リトライ

## Slack Webhook

### 概要
Slack Webhookを使用して、推薦記事をSlackチャンネルに投稿します。

### Webhook URLの取得

1. [Slack App Directory](https://slack.com/apps)にアクセス
2. 「Incoming Webhooks」を検索して選択
3. 「Add to Slack」をクリック
4. 投稿先のチャンネルを選択
5. Webhook URLをコピー

### 設定方法

#### config.ymlでの設定

```yaml
default_profile:
  output:
    slack_api:
      webhook_url: "https://hooks.slack.com/services/T00000000/B00000000/XXXXXXXXXXXXXXXXXXXX"
      # または
      webhook_url_env: "SLACK_WEBHOOK_URL"  # 環境変数名を指定（推奨）
      message_template: |
        :newspaper: *{{.Article.Title}}*
        {{.Article.Link}}
        
        {{.Comment}}
        
        {{if .FixedMessage}}{{.FixedMessage}}{{end}}
```

### メッセージテンプレート

利用可能な変数：
- `{{.Article.Title}}`: 記事タイトル
- `{{.Article.Link}}`: 記事URL
- `{{.Article.Description}}`: 記事概要
- `{{.Comment}}`: AIによるコメント
- `{{.FixedMessage}}`: 固定メッセージ

### Slackフォーマット

```yaml
message_template: |
  *{{.Article.Title}}*  # 太字
  _{{.Comment}}_  # イタリック
  `code`  # コード
  >引用  # 引用
  :emoji:  # 絵文字
```

### レート制限

- 1メッセージ/秒が推奨
- バースト時は短期間に複数送信可能
- 制限超過時は自動的に待機

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
      api_url: "https://misskey.io/api"  # インスタンスのAPI URL
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

### 対応インスタンス

- misskey.io
- misskey.dev
- その他のMisskeyインスタンス
- Firefish（Misskey派生）
- Foundkey（Misskey派生）

### 投稿オプション

```yaml
misskey:
  visibility: "public"  # public, home, followers, specified
  local_only: false  # ローカルタイムラインのみ
  cw: "長文注意"  # Content Warning
```

### レート制限

インスタンスごとに異なりますが、一般的な制限：
- 300ノート/時間
- 連続投稿は避ける（1秒以上の間隔推奨）

## 複数の出力先設定

### 同時投稿

```yaml
output:
  slack_api:
    webhook_url_env: "SLACK_WEBHOOK_URL"
    message_template: "{{.Article.Title}}\n{{.Article.Link}}"
  
  misskey:
    api_url: "https://misskey.io/api"
    api_token_env: "MISSKEY_API_TOKEN"
    message_template: "{{.Article.Title}}\n{{.Article.Link}}"
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

### 自動リトライ

ai-feedは以下の場合に自動リトライを実行：
- ネットワークエラー
- 一時的なサーバーエラー（5xx）
- レート制限（適切な待機時間後）

### リトライ設定

```yaml
# 将来的な設定例（現在は固定値）
retry:
  max_attempts: 3
  initial_delay: 1s
  max_delay: 30s
  exponential_backoff: true
```

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

### 一般的な問題と解決方法

#### Gemini API

**問題**: "API key not valid"
```bash
# APIキーの確認
echo $GEMINI_API_KEY

# 正しいAPIキーを設定
export GEMINI_API_KEY="correct-api-key"
```

**問題**: レート制限エラー
```yaml
# より低速なモデルに変更
ai:
  gemini:
    type: "gemini-2.5-flash-8b"  # 高速モデル
```

#### Slack

**問題**: "invalid_webhook_url"
```bash
# URLフォーマットを確認
# 正しい形式: https://hooks.slack.com/services/T.../B.../...
```

**問題**: メッセージが表示されない
```yaml
# テンプレートの構文を確認
message_template: |
  {{.Article.Title}}  # 正しい
  {{ .Article.Title }}  # スペースがあっても動作
```

#### Misskey

**問題**: "PERMISSION_DENIED"
```bash
# トークンの権限を確認
# write:notes 権限が必要
```

**問題**: APIエンドポイントエラー
```yaml
# URLの末尾を確認
api_url: "https://misskey.io/api"  # 正しい
# api_url: "https://misskey.io/api/"  # 末尾のスラッシュは不要
```

## ベストプラクティス

### 1. 環境変数の使用

```bash
# .env.exampleを作成
cat << EOF > .env.example
GEMINI_API_KEY=your-gemini-api-key
SLACK_WEBHOOK_URL=your-slack-webhook-url
MISSKEY_API_TOKEN=your-misskey-token
EOF

# 実際の.envファイルは.gitignoreに追加
echo ".env" >> .gitignore
```

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