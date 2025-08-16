package cmd

import (
	"github.com/spf13/cobra"
)

// version はビルド時にldflags で埋め込まれるバージョン情報
var version = "dev"

// makeVersionCmd はversionコマンドを生成する
func makeVersionCmd() *cobra.Command {
	cmd := &cobra.Command{
		Use:   "version",
		Short: "バージョン情報を表示",
		RunE: func(cmd *cobra.Command, args []string) error {
			cmd.Println(version)
			return nil
		},
	}
	return cmd
}
