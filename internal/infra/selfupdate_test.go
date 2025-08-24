package infra

import (
	"testing"
)

// TestNewSelfUpdater はSelfUpdaterの作成を確認する
func TestNewSelfUpdater(t *testing.T) {
	owner := "testowner"
	repo := "testrepo"

	updater := NewSelfUpdater(owner, repo)

	if updater == nil {
		t.Fatal("SelfUpdaterが作成されていません")
	}

	if updater.owner != owner {
		t.Errorf("期待されるownerは %q ですが、実際は %q でした", owner, updater.owner)
	}

	if updater.repo != repo {
		t.Errorf("期待されるrepoは %q ですが、実際は %q でした", repo, updater.repo)
	}

	if updater.releaseClient == nil {
		t.Error("GitHubReleaseClientが作成されていません")
	}
}

// TestGetCurrentVersion は現在のバージョン取得を確認する
func TestGetCurrentVersion(t *testing.T) {
	updater := NewSelfUpdater("owner", "repo")

	version, err := updater.GetCurrentVersion()
	if err != nil {
		t.Fatalf("現在のバージョン取得に失敗: %v", err)
	}

	// 何らかのバージョン文字列が返されることを確認
	if version == "" {
		t.Error("バージョンが空文字です")
	}

	// "dev" または実際のバージョン文字列が返されることを期待
	if version != "dev" && len(version) == 0 {
		t.Errorf("予期しないバージョン文字列: %q", version)
	}
}
