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

	t.Run("ビルド時バージョンがdevかつビルド情報にバージョンがある場合", func(t *testing.T) {
		originalVersion := version
		originalReadBuildInfo := readBuildInfo
		version = "dev"
		readBuildInfo = func() (*debug.BuildInfo, bool) {
			return &debug.BuildInfo{
				Main: debug.Module{
					Version: "v1.0.0",
				},
			}, true
		}
		defer func() {
			version = originalVersion
			readBuildInfo = originalReadBuildInfo
		}()

		got := getVersion()
		if got != "v1.0.0" {
			t.Errorf("期待される結果は v1.0.0 ですが、実際は %s でした", got)
		}
	})

	t.Run("ビルド時バージョンがdevかつビルド情報が(devel)の場合", func(t *testing.T) {
		originalVersion := version
		originalReadBuildInfo := readBuildInfo
		version = "dev"
		readBuildInfo = func() (*debug.BuildInfo, bool) {
			return &debug.BuildInfo{
				Main: debug.Module{
					Version: "(devel)",
				},
			}, true
		}
		defer func() {
			version = originalVersion
			readBuildInfo = originalReadBuildInfo
		}()

		got := getVersion()
		if got != "dev" {
			t.Errorf("期待される結果は dev ですが、実際は %s でした", got)
		}
	})

	t.Run("ビルド時バージョンがdevかつビルド情報が取得できない場合", func(t *testing.T) {
		originalVersion := version
		originalReadBuildInfo := readBuildInfo
		version = "dev"
		readBuildInfo = func() (*debug.BuildInfo, bool) {
			return nil, false
		}
		defer func() {
			version = originalVersion
			readBuildInfo = originalReadBuildInfo
		}()

		got := getVersion()
		if got != "dev" {
			t.Errorf("期待される結果は dev ですが、実際は %s でした", got)
		}
	})
}
