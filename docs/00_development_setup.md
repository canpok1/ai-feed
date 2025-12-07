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

#### 4. GitHub CLI（gh）

GitHub CLIは、issueやプルリクエストの操作に使用します。

- **インストール方法**:
  ```bash
  # macOS
  brew install gh

  # Linux (Ubuntu/Debian)
  type -p curl >/dev/null || (sudo apt update && sudo apt install curl -y)
  curl -fsSL https://cli.github.com/packages/githubcli-archive-keyring.gpg | sudo dd of=/usr/share/keyrings/githubcli-archive-keyring.gpg
  sudo chmod go+r /usr/share/keyrings/githubcli-archive-keyring.gpg
  echo "deb [arch=$(dpkg --print-architecture) signed-by=/usr/share/keyrings/githubcli-archive-keyring.gpg] https://cli.github.com/packages stable main" | sudo tee /etc/apt/sources.list.d/github-cli.list > /dev/null
  sudo apt update
  sudo apt install gh -y

  # Windows
  # https://cli.github.com/ からインストーラをダウンロード
  ```

- **認証方法**:

  **推奨: gh auth login を使用**
  ```bash
  gh auth login
  ```
  対話的に認証を行い、tokenを安全に保存します。

  **CI/CD環境: 環境変数を使用**
  ```bash
  export GH_TOKEN="your_token_here"
  ```

- **トークンの作成（手動で作成する場合）**:
  1. GitHub Settings > Developer settings > Personal access tokens > Tokens (classic)
  2. 以下の権限を付与:
     - `repo` (リポジトリへの完全アクセス)
     - `read:org` (組織情報の読み取り)

- **重要な注意事項**:
  - **絶対にトークンをコードやコミットに含めないでください**
  - .envファイルは.gitignoreに含まれており、リポジトリにコミットされません
  - CI/CD環境ではGitHub Secretsを使用してください
  - トークンは定期的にローテーションすることを推奨します


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

## 次のステップ

開発環境の構築が完了したら、以下のドキュメントも参照してください：

- [コーディングルール](./01_coding_rules.md) - コーディング規約とベストプラクティス
- [アーキテクチャ概要](./02_architecture.md) - プロジェクトの構造と設計
- [テストガイド](./03_testing.md) - テストの書き方と実行方法
- [コントリビューションガイド](./04_contributing.md) - 貢献方法とプルリクエストの作成
- [API連携ガイド](./05_api_integration.md) - 外部API設定の詳細
- [スクリプトガイド](./06_scripts.md) - 便利なスクリプトの使い方