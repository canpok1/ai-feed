package cmd

import (
	"fmt"

	"github.com/canpok1/ai-feed/internal/infra"
	"github.com/spf13/cobra"
)

func makeConfigCmd() *cobra.Command {
	cmd := &cobra.Command{
		Use:   "config",
		Short: "The config command manages config.yml: generates boilerplate or validates settings.",
		Run: func(cmd *cobra.Command, args []string) {
			fmt.Println("config called")
			fmt.Printf("config:%s\n", cfgFile)
		},
	}
	return cmd
}

const DefaultConfigFilePath = "./config.yml"

func makeConfigInitCmd() *cobra.Command {
	cmd := &cobra.Command{
		Use:   "init",
		Short: "Generates a boilerplate config.yml file. It will not overwrite an existing file.",
		RunE: func(cmd *cobra.Command, args []string) error {
			filePath := DefaultConfigFilePath

			configRepo := infra.NewYamlConfigRepository(filePath)

			config := infra.MakeDefaultConfig()
			if err := configRepo.Save(config); err != nil {
				return err
			}
			fmt.Printf("%s generated\n", filePath)
			return nil
		},
	}
	return cmd
}
