package version

import (
	"runtime/debug"
)

// GetVersion はバージョン情報を取得する
// ビルド時に設定されたバージョンを優先し、"dev"の場合はビルド情報から取得する
func GetVersion(buildTimeVersion string) string {
	if buildTimeVersion != "dev" {
		return buildTimeVersion
	}

	// go installでビルドされた場合、ビルド情報からバージョンを取得
	if info, ok := debug.ReadBuildInfo(); ok {
		if info.Main.Version != "(devel)" && info.Main.Version != "" {
			return info.Main.Version
		}
	}

	return "dev"
}

// GetVersionWithReadBuildInfo はテスト時にモック可能なバージョン取得関数
func GetVersionWithReadBuildInfo(buildTimeVersion string, readBuildInfoFunc func() (*debug.BuildInfo, bool)) string {
	if buildTimeVersion != "dev" {
		return buildTimeVersion
	}

	// go installでビルドされた場合、ビルド情報からバージョンを取得
	if info, ok := readBuildInfoFunc(); ok {
		if info.Main.Version != "(devel)" && info.Main.Version != "" {
			return info.Main.Version
		}
	}

	return "dev"
}
