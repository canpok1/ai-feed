package cmd

import (
	"bytes"
	"runtime/debug"
	"strings"
	"testing"
)

// TestVersionCommand はversionコマンドの動作を確認する
func TestVersionCommand(t *testing.T) {
	testCases := []struct {
		name    string
		version string
		want    string
	}{
		{
			name:    "default version",
			version: "dev",
			want:    "dev",
		},
		{
			name:    "custom version",
			version: "v1.2.3",
			want:    "v1.2.3",
		},
	}

	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			originalVersion := version
			version = tc.version
			defer func() { version = originalVersion }()

			var buf bytes.Buffer
			cmd := makeVersionCmd()
			cmd.SetOut(&buf)

			err := cmd.Execute()
			if err != nil {
				t.Fatalf("versionコマンドの実行に失敗: %v", err)
			}

			got := strings.TrimSpace(buf.String())
			if got != tc.want {
				t.Errorf("期待される出力は %q ですが、実際は %q でした", tc.want, got)
			}
		})
	}
}

// TestGetVersion はgetVersion関数の動作を確認する
func TestGetVersion(t *testing.T) {
	t.Run("ビルド時バージョンが設定されている場合", func(t *testing.T) {
		originalVersion := version
		version = "v1.2.3"
		defer func() { version = originalVersion }()

		got := getVersion()
		if got != "v1.2.3" {
			t.Errorf("期待される結果は v1.2.3 ですが、実際は %s でした", got)
		}
	})

	t.Run("ビルド時バージョンがdevの場合", func(t *testing.T) {
		originalVersion := version
		version = "dev"
		defer func() { version = originalVersion }()

		// ビルド情報がない場合は"dev"が返される
		got := getVersion()
		// この環境ではビルド情報が取得できない可能性があるため、
		// "dev"または実際のバージョンのどちらかが返されることを許容
		if got != "dev" && !strings.HasPrefix(got, "v") {
			t.Errorf("期待される結果は dev またはバージョン文字列ですが、実際は %s でした", got)
		}
	})
}

// TestGetVersionWithMockBuildInfo はgo:embedなどでビルド情報が取得できる場合のテスト
func TestGetVersionWithMockBuildInfo(t *testing.T) {
	// このテストは実際のビルド情報に依存するため、
	// ビルド環境での動作を確認するためのもの
	if info, ok := debug.ReadBuildInfo(); ok {
		t.Logf("BuildInfo.Main.Version: %s", info.Main.Version)
		t.Logf("BuildInfo.Main.Path: %s", info.Main.Path)
	} else {
		t.Log("BuildInfoが取得できません")
	}
}
