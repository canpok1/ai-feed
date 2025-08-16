package cmd

import (
	"runtime/debug"

	"github.com/spf13/cobra"
)

// version はビルド時にldflags で埋め込まれるバージョン情報
var version = "dev"

// getVersion はバージョン情報を取得する
// ビルド時に設定されたバージョンを優先し、"dev"の場合はビルド情報から取得する
func getVersion() string {
	if version != "dev" {
		return version
	}

	// go installでビルドされた場合、ビルド情報からバージョンを取得
	if info, ok := debug.ReadBuildInfo(); ok {
		if info.Main.Version != "(devel)" && info.Main.Version != "" {
			return info.Main.Version
		}
	}

	return "dev"
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
