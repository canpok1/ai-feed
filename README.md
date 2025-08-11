# ai-feed

AI Feedは、指定されたURLから記事をプレビューしたり、AIが記事を推薦したり、プロファイルを管理したりするためのCLIツールです。

## クイックスタート

1. [リリース一覧](https://github.com/canpok1/ai-feed/releases)から自分の環境に合ったものをダウンロードして任意の場所に解凍する。

2. configファイルを生成
    ```
    ./ai-feed config init
    ```

3. config.ymlを編集
    - geminiのtypeにはモデルのバージョンを指定。具体的な値は[参考リンク](https://ai.google.dev/gemini-api/docs/models?hl=ja#model-versions)を参照。
    - APIキー・トークンの設定方法は[APIキーの設定](#apiキーの設定)を参照。
    - 投稿文に固定文言を付与しないなら fixed_message の行を削除。
    - slackへの投稿を行わないなら slack_api のブロックを削除。
    - misskeyへの投稿を行わないなら misskey のブロックを削除。

4. おすすめ記事を表示
    ```
    ./ai-feed recommend --url {任意のRSSフィードのURL}
    ```

## 開発環境

### 必要条件

- golang: 1.24

### 各種操作

基本的にmakeコマンドで操作可能

```
# run（使い方確認）
make run

# run（オプション指定）
make run option="オプション"

# build
make build

# clean
make clean

# test
make test

# lint (静的解析)
make lint

# fmt (コードフォーマット)
make fmt
```

## コマンド

### `preview` コマンド

`preview` コマンドは、指定されたURLから記事を一時的に取得し、表示します。購読やキャッシュは行いません。

#### `--url` オプション

プレビューするフィードのURLを指定します。複数のURLを指定できます。

例:
```bash
make run option="preview --url https://example.com/feed.xml --url https://another.com/rss"
```

#### `--source` オプション

URLのリストを含むファイルを指定します。ファイルは1行に1つのURLを記述します。空行はスキップされ、不正なURLは警告が表示されてスキップされます。

例:
```bash
# urls.txt の内容:
# https://example.com/feed.xml
# https://another.com/rss
# invalid-url

make run option="preview --source urls.txt"
```

`--source` オプションと `--url` オプションを同時に使用することはできません。

### `recommend` コマンド

`recommend` コマンドは、指定されたURLまたはファイルから記事をフェッチし、ランダムに選択した1つの記事にAIによる紹介文を付けて推奨します。

#### `--url` オプション

推奨する記事のフィードURLを指定します。

例:
```bash
make run option="recommend --url https://example.com/feed.xml"
```

#### `--source` オプション

URLのリストを含むファイルを指定します。ファイルは1行に1つのURLを記述します。

例:
```bash
# urls.txt の内容:
# https://example.com/feed.xml
# https://another.com/rss

make run option="recommend --source urls.txt"
```

`--source` オプションと `--url` オプションを同時に使用することはできません。

#### `--profile` オプション

プロファイルファイルを指定します。AIの設定や出力設定などが含まれます。

例:
```bash
make run option="recommend --url https://example.com/feed.xml --profile my_profile.yml"
```

### `profile` コマンド

`profile` コマンドは、ユーザープロファイルを管理します。

#### `init` サブコマンド

新しいプロファイルファイルを指定されたパスに初期化します。ファイルが既に存在する場合はエラーになります。

例:
```bash
make run option="profile init my_profile.yml"
```

### `config` コマンド

`config` コマンドは、`config.yml` ファイルを管理します。

#### `init` サブコマンド

ボイラープレートの `config.yml` ファイルを生成します。既存のファイルは上書きしません。

例:
```bash
make run option="config init"
```

## テンプレート記法

profile.ymlの設定でメッセージテンプレートを定義する際に、記事情報を動的に埋め込むためのテンプレート記法を使用できます。

### 基本的な記法

従来の記法（ドット記法）と新しい別名記法の両方がサポートされています：

| データ | 従来記法 | 別名記法 |
|--------|----------|----------|
| 記事タイトル | `{{.Title}}` または `{{.Article.Title}}` | `{{TITLE}}` |
| 記事URL | `{{.Link}}` または `{{.Article.Link}}` | `{{URL}}` |
| 記事内容 | `{{.Content}}` または `{{.Article.Content}}` | `{{CONTENT}}` |
| AIコメント | `{{.Comment}}` | `{{COMMENT}}` |
| 固定メッセージ | `{{.FixedMessage}}` | `{{FIXED_MESSAGE}}` |

### 使用例

```yaml
# PromptConfig での使用例
comment_prompt_template: |
  以下の記事について簡潔に紹介してください：
  タイトル: {{TITLE}}
  URL: {{URL}}
  内容: {{CONTENT}}

# SlackAPI での使用例
output:
  slack_api:
    message_template: |
      📰 おすすめ記事
      {{TITLE}}
      {{URL}}
      {{if .Comment}}
      💬 {{COMMENT}}
      {{end}}
```

### 注意事項

- **別名記法**: 大文字のみ使用可能（例: `{{TITLE}}`は正しいが、`{{title}}`や`{{Title}}`はエラー）
- **後方互換性**: 既存の記法も引き続き使用可能
- **混在可能**: 同一テンプレート内で新旧記法を混在させることも可能
- **条件分岐**: `{{if .Comment}}`などの制御構文も使用可能

## APIキーの設定

AI FeedでAPIキーやトークンを設定する方法は2つあります：

### 方法1: ファイルに直接記述

config.ymlやprofile.ymlファイルに直接APIキーを記述する方法です。

```yaml
# config.ymlの例
ai:
  gemini:
    api_key: your_actual_gemini_api_key_here

output:
  slack_api:
    api_token: your_actual_slack_token_here
  misskey:
    api_token: your_actual_misskey_token_here
```

### 方法2: 環境変数から取得（推奨）

セキュリティ上の理由から、APIキーを環境変数から取得する方法を推奨します。

```yaml
# config.ymlまたはprofile.ymlの例
ai:
  gemini:
    api_key_env: GEMINI_API_KEY  # 環境変数名を指定

output:
  slack_api:
    api_token_env: SLACK_TOKEN  # 環境変数名を指定
  misskey:
    api_token_env: MISSKEY_TOKEN  # 環境変数名を指定
```

環境変数の設定例：
```bash
# 環境変数を設定
export GEMINI_API_KEY="your_actual_gemini_api_key_here"
export SLACK_TOKEN="your_actual_slack_token_here"
export MISSKEY_TOKEN="your_actual_misskey_token_here"

# ai-feedを実行
./ai-feed recommend --url https://example.com/feed.xml
```

### 優先順位

両方の設定方法を併用した場合の優先順位：

1. **直接指定が最優先**: `api_key`や`api_token`が設定されている場合、環境変数の設定は無視されます
2. **環境変数**: 直接指定がない場合、環境変数から取得します
3. **エラー**: 直接指定も環境変数もない場合はエラーになります

### エラー対応

環境変数が設定されていない場合、以下のようなエラーが表示されます：

```
Profile validation failed:
  ERROR: 環境変数 'GEMINI_API_KEY' が設定されていません。ai.gemini.api_key_env で指定された環境変数を設定してください。
```

このエラーが発生した場合は、指定された環境変数名で正しくAPIキーが設定されているかを確認してください。
