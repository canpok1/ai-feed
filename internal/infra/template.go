package infra

// ProfileTemplateCore はプロファイル設定の共通部分のテンプレート文字列
const ProfileTemplateCore = `# AI設定
ai:
  gemini:
    type: gemini-2.5-flash                  # 使用するGeminiモデル
    api_key: xxxxxx                         # Google AI Studio APIキー

# プロンプト設定
system_prompt: あなたはXXXXなAIアシスタントです。    # AIに与えるシステムプロンプト
comment_prompt_template: |                         # 記事紹介文生成用のプロンプトテンプレート
  以下の記事の紹介文を100字以内で作成してください。
  ---
  記事タイトル: {{title}}
  記事URL: {{url}}
  記事内容:
  {{content}}
fixed_message: 固定の文言です。                     # 記事紹介文に追加する固定文言

# 出力先設定
output:
  # Slack投稿設定
  slack_api:
    api_token: xxxxxx                       # Slack Bot Token
    api_url: https://example.com            # Slack API URL
    channel: "#general"                     # 投稿先チャンネル
  
  # Misskey投稿設定
  misskey:
    api_token: xxxxxx                       # Misskeyアクセストークン
    api_url: https://misskey.social/api     # MisskeyのAPIエンドポイント`
