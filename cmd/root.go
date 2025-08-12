package cmd

import (
	"github.com/canpok1/ai-feed/internal/infra/fetch"
	"github.com/spf13/cobra"
)

var cfgFile string

func makeRootCmd() *cobra.Command {
	cmd := &cobra.Command{
		Use:   "ai-feed",
		Short: "An AI-powered CLI RSS reader that summarizes articles and posts comments to various platforms.",
	}
	cmd.PersistentFlags().StringVar(&cfgFile, "config", "", "config file (default is ./config.yml)")
	cmd.SilenceUsage = true
	return cmd
}

func Execute() error {
	rootCmd := makeRootCmd()

	recommendCmd := makeRecommendCmd(fetch.NewFetchClient())
	rootCmd.AddCommand(recommendCmd)

	rootCmd.AddCommand(makeInitCmd())

	rootCmd.AddCommand(makeProfileCmd())

	return rootCmd.Execute()
}
