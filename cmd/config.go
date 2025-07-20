package cmd

import (
	"fmt"

	"github.com/spf13/cobra"
)

func makeConfigCmd() *cobra.Command {
	cmd := &cobra.Command{
		Use:   "config",
		Short: "The config command manages config.yml: generates boilerplate or validates settings.",
		Run: func(cmd *cobra.Command, args []string) {
			fmt.Println("config called")
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
