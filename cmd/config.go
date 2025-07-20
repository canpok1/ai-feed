package cmd

import (
	"fmt"

	"github.com/canpok1/ai-feed/internal/domain/entity"
	"github.com/canpok1/ai-feed/internal/infra"
	"github.com/spf13/cobra"
)

func makeConfigCmd() *cobra.Command {
	cmd := &cobra.Command{
		Use:   "config",
		Short: "The config command manages config.yml: generates boilerplate or validates settings.",
		Run: func(cmd *cobra.Command, args []string) {
			fmt.Println("config called")
			fmt.Printf("config:%s", cfgFile)
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

			config := entity.MakeDefaultConfig()
			if err := configRepo.Save(config); err != nil {
				fmt.Println(err)
			} else {
				fmt.Printf("%s generated\n", filePath)
			}
		},
	}
	return cmd
}

var configCmd = makeConfigCmd()
var configInitCmd = makeConfigInitCmd()

func init() {
	rootCmd.AddCommand(configCmd)
	configCmd.AddCommand(configInitCmd)
}
