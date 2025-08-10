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
		Short: "Generates a boilerplate config.yml file. It will not overwrite an existing file.",
		RunE: func(cmd *cobra.Command, args []string) error {
			filePath := DefaultConfigFilePath

			configRepo := infra.NewYamlConfigRepository(filePath)

			// テンプレートを使用してコメント付きconfig.ymlを生成
			if err := configRepo.SaveWithTemplate(); err != nil {
				return err
			}
			fmt.Printf("%s generated\n", filePath)
			return nil
		},
	}
	return cmd
}
