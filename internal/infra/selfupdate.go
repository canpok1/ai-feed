package infra

import (
	"context"
	"fmt"
	"os"

	"github.com/canpok1/ai-feed/internal/domain"
	versionpkg "github.com/canpok1/ai-feed/internal/version"
	"github.com/creativeprojects/go-selfupdate"
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
	// 共通のバージョン取得ロジックを使用
	return versionpkg.GetVersion("dev"), nil
}

// GetLatestVersion は最新のリリース情報を取得する
func (s *SelfUpdater) GetLatestVersion() (*domain.ReleaseInfo, error) {
	return s.releaseClient.GetLatestRelease()
}

// UpdateBinary はリリース情報を使用してバイナリを更新する
func (s *SelfUpdater) UpdateBinary(latest *domain.ReleaseInfo) error {
	ctx := context.Background()

	// 実行ファイルのパスを取得
	executable, err := os.Executable()
	if err != nil {
		return fmt.Errorf("実行ファイルのパスを取得できません: %w", err)
	}

	// 現在の実行ファイルを更新
	err = selfupdate.UpdateTo(ctx, latest.AssetURL, executable, "ai-feed")
	if err != nil {
		// 権限エラーやファイルシステムエラーの処理
		if os.IsPermission(err) {
			return fmt.Errorf("更新に必要な権限がありません。管理者権限で実行してください: %w", err)
		}
		return fmt.Errorf("バイナリの更新に失敗しました: %w", err)
	}

	return nil
}
