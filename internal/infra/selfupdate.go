package infra

import (
	"fmt"
	"os"
	"runtime/debug"

	"github.com/canpok1/ai-feed/internal/domain"
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
	// 実行中のバイナリファイルのパスを取得
	executable, err := os.Executable()
	if err != nil {
		return fmt.Errorf("実行ファイルのパスを取得できません: %w", err)
	}

	// updaterを初期化
	updater, err := selfupdate.NewUpdater(selfupdate.Config{
		Source: selfupdate.NewGitHubSource(selfupdate.GitHubConfig{
			RepositoryOwner: s.owner,
			RepositoryName:  s.repo,
		}),
		Executable: executable,
	})
	if err != nil {
		return fmt.Errorf("updaterの初期化に失敗しました: %w", err)
	}

	// 更新の実行
	err = updater.UpdateTo(version)
	if err != nil {
		// 権限エラーやファイルシステムエラーの処理
		if os.IsPermission(err) {
			return fmt.Errorf("更新に必要な権限がありません。管理者権限で実行してください: %w", err)
		}
		return fmt.Errorf("バイナリの更新に失敗しました: %w", err)
	}

	return nil
}
