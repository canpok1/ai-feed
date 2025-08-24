package infra

import (
	"runtime"
	"strings"
	"testing"

	"github.com/google/go-github/v65/github"
)

// TestNewGitHubReleaseClient はGitHubReleaseClientの作成を確認する
func TestNewGitHubReleaseClient(t *testing.T) {
	owner := "testowner"
	repo := "testrepo"

	client := NewGitHubReleaseClient(owner, repo)

	if client == nil {
		t.Fatal("GitHubReleaseClientが作成されていません")
	}

	if client.owner != owner {
		t.Errorf("期待されるownerは %q ですが、実際は %q でした", owner, client.owner)
	}

	if client.repo != repo {
		t.Errorf("期待されるrepoは %q ですが、実際は %q でした", repo, client.repo)
	}

	if client.client == nil {
		t.Error("GitHubクライアントが作成されていません")
	}
}

// TestSelectAsset はOS/アーキテクチャに応じたアセット選択を確認する
func TestSelectAsset(t *testing.T) {
	client := NewGitHubReleaseClient("owner", "repo")

	// テスト用のアセットを作成
	assets := []*github.ReleaseAsset{
		{
			Name:               github.String("ai-feed_Linux_x86_64.tar.gz"),
			BrowserDownloadURL: github.String("https://example.com/linux-amd64"),
		},
		{
			Name:               github.String("ai-feed_Darwin_arm64.tar.gz"),
			BrowserDownloadURL: github.String("https://example.com/darwin-arm64"),
		},
		{
			Name:               github.String("ai-feed_Windows_x86_64.zip"),
			BrowserDownloadURL: github.String("https://example.com/windows-amd64"),
		},
	}

	// 現在のOS/アーキテクチャに応じた期待値を設定
	var expectedURL string
	currentOS := runtime.GOOS
	currentArch := runtime.GOARCH

	switch {
	case currentOS == "linux" && currentArch == "amd64":
		expectedURL = "https://example.com/linux-amd64"
	case currentOS == "darwin" && currentArch == "arm64":
		expectedURL = "https://example.com/darwin-arm64"
	case currentOS == "windows" && currentArch == "amd64":
		expectedURL = "https://example.com/windows-amd64"
	default:
		expectedURL = "" // 該当するアセットがない場合
	}

	result := client.selectAsset(assets)

	if result != expectedURL {
		t.Errorf("期待されるURLは %q ですが、実際は %q でした (OS: %s, Arch: %s)",
			expectedURL, result, currentOS, currentArch)
	}
}

// TestSelectAssetNoMatch は該当するアセットがない場合を確認する
func TestSelectAssetNoMatch(t *testing.T) {
	client := NewGitHubReleaseClient("owner", "repo")

	// 現在のOS/アーキテクチャに該当しないアセットのみを作成
	assets := []*github.ReleaseAsset{
		{
			Name:               github.String("ai-feed_UnknownOS_UnknownArch.tar.gz"),
			BrowserDownloadURL: github.String("https://example.com/unknown"),
		},
	}

	result := client.selectAsset(assets)

	if result != "" {
		t.Errorf("該当するアセットがない場合は空文字を返すべきですが、実際は %q でした", result)
	}
}

// TestSelectAssetEmptyList は空のアセットリストの場合を確認する
func TestSelectAssetEmptyList(t *testing.T) {
	client := NewGitHubReleaseClient("owner", "repo")

	var assets []*github.ReleaseAsset

	result := client.selectAsset(assets)

	if result != "" {
		t.Errorf("空のアセットリストの場合は空文字を返すべきですが、実際は %q でした", result)
	}
}

// TestAssetNameMatching はアセット名のマッチングロジックを確認する
func TestAssetNameMatching(t *testing.T) {
	_ = NewGitHubReleaseClient("owner", "repo")

	testCases := []struct {
		name      string
		assetName string
		os        string
		arch      string
		expected  bool
	}{
		{
			name:      "Linux x86_64マッチ",
			assetName: "ai-feed_Linux_x86_64.tar.gz",
			os:        "linux",
			arch:      "amd64",
			expected:  true,
		},
		{
			name:      "Darwin arm64マッチ",
			assetName: "ai-feed_Darwin_arm64.tar.gz",
			os:        "darwin",
			arch:      "arm64",
			expected:  true,
		},
		{
			name:      "Windows x86_64マッチ",
			assetName: "ai-feed_Windows_x86_64.zip",
			os:        "windows",
			arch:      "amd64",
			expected:  true,
		},
		{
			name:      "OSミスマッチ",
			assetName: "ai-feed_Linux_x86_64.tar.gz",
			os:        "windows",
			arch:      "amd64",
			expected:  false,
		},
		{
			name:      "アーキテクチャミスマッチ",
			assetName: "ai-feed_Linux_arm64.tar.gz",
			os:        "linux",
			arch:      "amd64",
			expected:  false,
		},
	}

	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			// OSとアーキテクチャのマッピング
			osMap := map[string]string{
				"linux":   "Linux",
				"darwin":  "Darwin",
				"windows": "Windows",
			}
			archMap := map[string]string{
				"amd64": "x86_64",
				"arm64": "arm64",
				"386":   "i386",
			}

			mappedOS := osMap[tc.os]
			mappedArch := archMap[tc.arch]

			result := strings.Contains(tc.assetName, mappedOS) && strings.Contains(tc.assetName, mappedArch)

			if result != tc.expected {
				t.Errorf("アセット名 %q のマッチング結果が期待値と異なります (OS: %s, Arch: %s, 期待値: %v, 実際: %v)",
					tc.assetName, tc.os, tc.arch, tc.expected, result)
			}
		})
	}
}
