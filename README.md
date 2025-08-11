# ai-feed

AI Feedは指定RSSフィードからランダム選択した記事をAIの感想付きで表示するCLIツールです。

## クイックスタート

1. [リリース一覧](https://github.com/canpok1/ai-feed/releases)から自分の環境に合ったものをダウンロードして任意の場所に解凍する。

2. configファイルを生成
    ```
    ./ai-feed init
    ```

3. config.ymlを編集
    - 編集が必要なものは次の通り
        - default_profile.ai.gemini
            - api_key もしくは api_key_env
                - 有効なAPIキーを設定
        - default_profile.system_prompt
            - オリジナルの設定に変更
        - default_profile.fixed_message
            - 固定で埋め込みたい文言があれば設定、埋め込む必要がないなら項目削除
        - default_profile.output.slack_api
            - Slack投稿しない場合は項目削除
            - Slack投稿するなら下記項目を設定
                - api_token もしくは api_token_env
                - channel
        - default_profile.output.misskey
            - Misskey投稿しない場合は項目削除
            - Misskey投稿するなら下記項目を設定
                - api_token もしくは api_token_env
                - api_url

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
