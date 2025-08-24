package infra

import (
	"context"
	"fmt"
	"runtime"
	"strings"

	"github.com/canpok1/ai-feed/internal/domain"
	"github.com/google/go-github/v30/github"
)

// GitHubReleaseClient はGitHubのリリース情報を取得するクライアント
type GitHubReleaseClient struct {
	owner  string
	repo   string
	client *github.Client
}

// NewGitHubReleaseClient は新しいGitHubReleaseClientを作成する
func NewGitHubReleaseClient(owner, repo string) *GitHubReleaseClient {
	return &GitHubReleaseClient{
		owner:  owner,
		repo:   repo,
		client: github.NewClient(nil),
	}
}

// GetLatestRelease は最新の安定版リリース情報を取得する
func (g *GitHubReleaseClient) GetLatestRelease() (*domain.ReleaseInfo, error) {
	ctx := context.Background()

	// 最新のリリースを取得（プレリリース版を除外）
	release, _, err := g.client.Repositories.GetLatestRelease(ctx, g.owner, g.repo)
	if err != nil {
		return nil, fmt.Errorf("最新リリースの取得に失敗しました: %w", err)
	}

	// プレリリース版の場合はスキップ
	if release.GetPrerelease() {
		return nil, fmt.Errorf("最新リリースがプレリリース版です")
	}

	// OS/アーキテクチャに応じたアセットを選択
	assetURL := g.selectAsset(release.Assets)
	if assetURL == "" {
		return nil, fmt.Errorf("対応するバイナリが見つかりません: OS=%s, Arch=%s", runtime.GOOS, runtime.GOARCH)
	}

	return &domain.ReleaseInfo{
		Version:      release.GetTagName(),
		AssetURL:     assetURL,
		ReleaseNotes: release.GetBody(),
	}, nil
}

// selectAsset はOS/アーキテクチャに応じた適切なアセットを選択する
func (g *GitHubReleaseClient) selectAsset(assets []*github.ReleaseAsset) string {
	// アセット名のパターンを構築
	// 例: ai-feed_Linux_x86_64.tar.gz, ai-feed_Darwin_arm64.tar.gz
	osName := runtime.GOOS
	archName := runtime.GOARCH

	// Goのランタイム名をリリースアセット名に変換
	osMap := map[string]string{
		"darwin":  "Darwin",
		"linux":   "Linux",
		"windows": "Windows",
	}

	archMap := map[string]string{
		"amd64": "x86_64",
		"arm64": "arm64",
		"386":   "i386",
	}

	if mappedOS, ok := osMap[osName]; ok {
		osName = mappedOS
	}

	if mappedArch, ok := archMap[archName]; ok {
		archName = mappedArch
	}

	// アセットから適切なものを選択
	for _, asset := range assets {
		name := asset.GetName()
		// OS名とアーキテクチャ名が含まれているかチェック
		if strings.Contains(name, osName) && strings.Contains(name, archName) {
			return asset.GetBrowserDownloadURL()
		}
	}

	// 見つからない場合は空文字を返す
	return ""
}

