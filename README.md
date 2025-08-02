# ai-feed

AI Feedは、指定されたURLから記事をプレビューしたり、AIが記事を推薦したり、プロファイルを管理したりするためのCLIツールです。

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

`recommend` コマンドは、指定されたURLまたはファイルから記事をフェッチし、AIがその中からランダムに1つの記事を推奨します。

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
