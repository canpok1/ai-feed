# 開発環境セットアップガイド

このドキュメントでは、ai-feedプロジェクトの開発環境を構築する手順を説明します。

## 前提条件

### 必須ツール

#### 1. Go言語
- **バージョン**: 1.24以上
- **インストール方法**:
  ```bash
  # macOS (Homebrewを使用)
  brew install go
  
  # Linux (公式インストーラを使用)
  wget https://go.dev/dl/go1.24.linux-amd64.tar.gz
  sudo rm -rf /usr/local/go && sudo tar -C /usr/local -xzf go1.24.linux-amd64.tar.gz
  export PATH=$PATH:/usr/local/go/bin
  
  # Windows
  # https://go.dev/dl/ から.msiインストーラをダウンロードして実行
  ```

- **確認方法**:
  ```bash
  go version
  # go version go1.24 ...
  ```

#### 2. Make
- **インストール方法**:
  ```bash
  # macOS (通常はプリインストール済み、ない場合)
  xcode-select --install
  
  # Linux (Ubuntu/Debian)
  sudo apt-get update
  sudo apt-get install build-essential
  
  # Linux (CentOS/RHEL/Fedora)
  sudo yum groupinstall "Development Tools"
  
  # Windows (Git Bashに含まれる、またはChocolateyを使用)
  choco install make
  ```

#### 3. Git
- **インストール方法**:
  ```bash
  # macOS
  brew install git
  
  # Linux (Ubuntu/Debian)
  sudo apt-get install git
  
  # Windows
  # https://git-scm.com/ からインストーラをダウンロード
  ```

### 推奨ツール

#### GitHub CLI (gh)
プルリクエストやIssueの管理に便利です。

```bash
# macOS
brew install gh

# Linux
# https://github.com/cli/cli/blob/trunk/docs/install_linux.md を参照

# Windows
winget install --id GitHub.cli
```

## プロジェクトのセットアップ

### 1. リポジトリのクローン
```bash
git clone https://github.com/canpok1/ai-feed.git
cd ai-feed
```

### 2. 開発依存関係のインストール
```bash
# 開発に必要なツールをインストール
make setup
```

このコマンドは以下を実行します：
- Go モジュールの依存関係をダウンロード
- 静的解析ツール（golangci-lint）のインストール
- モック生成ツール（mockgen）のインストール

### 3. 動作確認
```bash
# ビルドが成功することを確認
make build

# テストが通ることを確認
make test

# ヘルプを表示
./ai-feed --help
```

## 開発用コマンド

### ビルド関連

```bash
# アプリケーションのビルド
make build

# ビルド成果物のクリーンアップ
make clean

# アプリケーションの実行（使い方確認）
make run

# オプション付きで実行
make run option="recommend --url https://example.com/feed.rss"
```

### テスト関連

```bash
# 全テストの実行
make test

# 特定のパッケージのテストを実行
go test ./internal/domain/...

# 詳細な出力でテストを実行
go test -v ./...

# カバレッジレポートの生成
go test -coverprofile=coverage.out ./...
go tool cover -html=coverage.out
```

### コード品質

```bash
# コードのフォーマット
make fmt

# 静的解析（コミット前に必ず実行）
make lint

# モックの生成（インターフェース変更後）
make generate
```

### 依存関係管理

```bash
# 依存関係の更新
go get -u ./...

# 不要な依存関係の削除
go mod tidy

# 依存関係の確認
go mod graph
```

## IDE/エディタの設定

### Visual Studio Code

推奨拡張機能：
- [Go](https://marketplace.visualstudio.com/items?itemName=golang.Go)
- [GitLens](https://marketplace.visualstudio.com/items?itemName=eamodio.gitlens)
- [YAML](https://marketplace.visualstudio.com/items?itemName=redhat.vscode-yaml)

`.vscode/settings.json`の推奨設定：
```json
{
  "go.lintTool": "golangci-lint",
  "go.lintOnSave": "workspace",
  "go.formatTool": "goimports",
  "go.testOnSave": true,
  "[go]": {
    "editor.formatOnSave": true,
    "editor.codeActionsOnSave": {
      "source.organizeImports": true
    }
  }
}
```

### GoLand/IntelliJ IDEA

1. File → Settings → Go → Go Modules で「Enable Go modules integration」を有効化
2. File → Settings → Tools → File Watchers で以下を設定：
   - goimports（保存時の自動フォーマット）
   - golangci-lint（保存時の静的解析）

### Vim/Neovim

推奨プラグイン：
- [vim-go](https://github.com/fatih/vim-go)
- [coc.nvim](https://github.com/neoclide/coc.nvim) with coc-go

`.vimrc`または`init.vim`の設定例：
```vim
let g:go_fmt_command = "goimports"
let g:go_metalinter_enabled = ['vet', 'golint', 'errcheck']
let g:go_metalinter_autosave = 1
```

## トラブルシューティング

### `make setup`が失敗する場合

1. Go のバージョンを確認
   ```bash
   go version
   ```

2. GOPATHとGO111MODULEの設定を確認
   ```bash
   go env GOPATH
   go env GO111MODULE
   ```

3. プロキシ環境下の場合
   ```bash
   export GOPROXY=https://proxy.golang.org,direct
   export GOSUMDB=sum.golang.org
   ```

### テストが失敗する場合

1. 依存関係を最新にする
   ```bash
   go mod download
   go mod tidy
   ```

2. モックを再生成する
   ```bash
   make generate
   ```

3. キャッシュをクリアする
   ```bash
   go clean -testcache
   ```

### `make lint`でエラーが出る場合

1. golangci-lintを最新バージョンに更新
   ```bash
   curl -sSfL https://raw.githubusercontent.com/golangci/golangci-lint/master/install.sh | sh -s -- -b $(go env GOPATH)/bin latest
   ```

2. 特定のルールを無効化したい場合は`.golangci.yml`を編集

## 次のステップ

開発環境の構築が完了したら、以下のドキュメントも参照してください：

- [コーディングルール](./01_coding_rules.md) - コーディング規約とベストプラクティス
- [アーキテクチャ概要](./02_architecture.md) - プロジェクトの構造と設計
- [テストガイド](./03_testing.md) - テストの書き方と実行方法
- [コントリビューションガイド](./04_contributing.md) - 貢献方法とプルリクエストの作成
- [API連携ガイド](./05_api_integration.md) - 外部API設定の詳細