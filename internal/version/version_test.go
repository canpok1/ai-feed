package version

import (
	"runtime/debug"
	"testing"
)

// TestGetVersion はGetVersion関数のテストを行う
func TestGetVersion(t *testing.T) {
	t.Run("ビルド時バージョンが設定されている場合", func(t *testing.T) {
		result := GetVersion("v1.2.3")
		if result != "v1.2.3" {
			t.Errorf("期待される結果は v1.2.3 ですが、実際は %s でした", result)
		}
	})

	t.Run("ビルド時バージョンがdevの場合", func(t *testing.T) {
		result := GetVersion("dev")
		// ビルド情報から取得されるか、"dev"が返される
		if result == "" {
			t.Error("結果が空文字です")
		}
	})
}

// TestGetVersionWithReadBuildInfo はGetVersionWithReadBuildInfo関数のテストを行う
func TestGetVersionWithReadBuildInfo(t *testing.T) {
	t.Run("ビルド時バージョンがdevかつビルド情報にバージョンがある場合", func(t *testing.T) {
		mockReadBuildInfo := func() (*debug.BuildInfo, bool) {
			return &debug.BuildInfo{
				Main: debug.Module{
					Version: "v1.0.0",
				},
			}, true
		}

		result := GetVersionWithReadBuildInfo("dev", mockReadBuildInfo)
		if result != "v1.0.0" {
			t.Errorf("期待される結果は v1.0.0 ですが、実際は %s でした", result)
		}
	})

	t.Run("ビルド時バージョンがdevかつビルド情報が(devel)の場合", func(t *testing.T) {
		mockReadBuildInfo := func() (*debug.BuildInfo, bool) {
			return &debug.BuildInfo{
				Main: debug.Module{
					Version: "(devel)",
				},
			}, true
		}

		result := GetVersionWithReadBuildInfo("dev", mockReadBuildInfo)
		if result != "dev" {
			t.Errorf("期待される結果は dev ですが、実際は %s でした", result)
		}
	})

	t.Run("ビルド時バージョンがdevかつビルド情報が取得できない場合", func(t *testing.T) {
		mockReadBuildInfo := func() (*debug.BuildInfo, bool) {
			return nil, false
		}

		result := GetVersionWithReadBuildInfo("dev", mockReadBuildInfo)
		if result != "dev" {
			t.Errorf("期待される結果は dev ですが、実際は %s でした", result)
		}
	})
}
