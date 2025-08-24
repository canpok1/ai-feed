package cmd

import (
	"fmt"

	"github.com/canpok1/ai-feed/internal/infra"
	"github.com/spf13/cobra"
)

const DefaultConfigFilePath = "./config.yml"

func makeInitCmd() *cobra.Command {
	cmd := &cobra.Command{
		Use:   "init",
		Short: "設定ファイル（config.yml）のテンプレートを生成します（既存ファイルは上書きしません）",
		RunE: func(cmd *cobra.Command, args []string) error {
			filePath := DefaultConfigFilePath

			fmt.Fprintln(cmd.ErrOrStderr(), "設定ファイルを初期化しています...")

			configRepo := infra.NewYamlConfigRepository(filePath)

			fmt.Fprintln(cmd.ErrOrStderr(), "設定テンプレートを生成しています...")
			// テンプレートを使用してコメント付きconfig.ymlを生成
			if err := configRepo.SaveWithTemplate(); err != nil {
				return err
			}
			fmt.Printf("%s を生成しました\n", filePath)
			return nil
		},
	}
	cmd.SilenceUsage = true
	return cmd
}
