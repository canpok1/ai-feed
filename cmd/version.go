package cmd

import (
	"runtime/debug"

	versionpkg "github.com/canpok1/ai-feed/internal/version"
	"github.com/spf13/cobra"
)

// version はビルド時にldflags で埋め込まれるバージョン情報
var version = "dev"

// readBuildInfo はdebug.ReadBuildInfoを保持する変数（テスト時にモック可能）
var readBuildInfo = debug.ReadBuildInfo

// getVersion はバージョン情報を取得する
// ビルド時に設定されたバージョンを優先し、"dev"の場合はビルド情報から取得する
func getVersion() string {
	return versionpkg.GetVersionWithReadBuildInfo(version, readBuildInfo)
}

// makeVersionCmd はversionコマンドを生成する
func makeVersionCmd() *cobra.Command {
	cmd := &cobra.Command{
		Use:   "version",
		Short: "バージョン情報を表示",
		RunE: func(cmd *cobra.Command, args []string) error {
			cmd.Println(getVersion())
			return nil
		},
	}
	return cmd
}
