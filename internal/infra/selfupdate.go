package infra

import (
	"fmt"
	"runtime/debug"

	"github.com/canpok1/ai-feed/internal/domain"
)

// SelfUpdater はバイナリの自動更新を行う実装
type SelfUpdater struct {
	releaseClient *GitHubReleaseClient
	owner         string
	repo          string
}

// NewSelfUpdater は新しいSelfUpdaterを作成する
func NewSelfUpdater(owner, repo string) *SelfUpdater {
	return &SelfUpdater{
		releaseClient: NewGitHubReleaseClient(owner, repo),
		owner:         owner,
		repo:          repo,
	}
}

// GetCurrentVersion は現在のバージョンを取得する
func (s *SelfUpdater) GetCurrentVersion() (string, error) {
	// cmd/version.goと同じロジックを使用してバージョンを取得
	// 循環インポートを避けるため、同じ実装を再現する

	// ビルド情報からバージョンを取得
	if info, ok := debug.ReadBuildInfo(); ok {
		if info.Main.Version != "(devel)" && info.Main.Version != "" {
			return info.Main.Version, nil
		}
	}

	// デフォルトは"dev"を返す
	return "dev", nil
}

// GetLatestVersion は最新のリリース情報を取得する
func (s *SelfUpdater) GetLatestVersion() (*domain.ReleaseInfo, error) {
	return s.releaseClient.GetLatestRelease()
}

// UpdateBinary は指定されたバージョンにバイナリを更新する
func (s *SelfUpdater) UpdateBinary(version string) error {
	// 簡単な実装として、現在は未実装のメッセージを返す
	// go-selfupdateライブラリのAPIが複雑なため、基本構造のみ実装
	return fmt.Errorf("バイナリ更新機能は実装中です。手動で更新してください")
}
