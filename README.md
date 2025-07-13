# ai-feed

AI Feedは、指定されたURLから記事をプレビューしたり、RSSフィードを購読したりするためのCLIツールです。

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

# vet (静的解析)
make vet

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
