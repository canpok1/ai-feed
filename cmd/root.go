package cmd

import (
	"github.com/canpok1/ai-feed/internal/domain"
	"github.com/canpok1/ai-feed/internal/infra"
	"github.com/spf13/cobra"
)

var cfgFile string

func makeRootCmd() *cobra.Command {
	cmd := &cobra.Command{
		Use:   "ai-feed",
		Short: "An AI-powered CLI RSS reader that summarizes articles and posts comments to various platforms.",
	}
	cmd.PersistentFlags().StringVar(&cfgFile, "config", "", "config file (default is ./config.yml)")
	return cmd
}

func Execute() error {
	rootCmd := makeRootCmd()

	previewCmd := makePreviewCmd()
	rootCmd.AddCommand(previewCmd)

	recommendCmd := makeRecommendCmd(infra.NewFetchClient(), domain.NewRandomRecommender(infra.NewCommentGeneratorFactory()))
	rootCmd.AddCommand(recommendCmd)

	configCmd := makeConfigCmd()
	configInitCmd := makeConfigInitCmd()
	configCmd.AddCommand(configInitCmd)
	rootCmd.AddCommand(configCmd)

	rootCmd.AddCommand(makeProfileCmd())

	return rootCmd.Execute()
}
