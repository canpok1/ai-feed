package cmd

import (
	"github.com/canpok1/ai-feed/internal/infra/fetch"
	"github.com/spf13/cobra"
)

var cfgFile string

func makeRootCmd() *cobra.Command {
	cmd := &cobra.Command{
		Use:   "ai-feed",
		Short: "RSSフィードから記事を取得し、AIによる要約とコメント投稿を行うCLIツールです",
	}
	cmd.PersistentFlags().StringVar(&cfgFile, "config", "", "設定ファイル (デフォルトは ./config.yml)")
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
