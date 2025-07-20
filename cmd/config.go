package cmd

import (
	"fmt"

	"github.com/canpok1/ai-feed/internal/domain"
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

func makeConfigInitCmd() *cobra.Command {
	cmd := &cobra.Command{
		Use:   "init",
		Short: "Generates a boilerplate config.yml file, prompting for overwrite if it exists.",
		Run: func(cmd *cobra.Command, args []string) {
			fmt.Println("config init called")

			config := domain.MakeDefaultConfig()
			configRepo := infra.NewYamlConfigRepository(&cfgFile)
			if err := configRepo.Save(config); err != nil {
				fmt.Println(err)
			} else {
				fmt.Println("config.yml generated")
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
