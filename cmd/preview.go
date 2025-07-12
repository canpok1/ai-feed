package cmd

import (
	"fmt"

	"github.com/canpok1/ai-feed/internal"
	"github.com/spf13/cobra"
)

var previewCmd = &cobra.Command{
	Use:   "preview",
	Short: "The preview command temporarily fetches and displays articles from specified URLs or files without subscribing or caching them.",
	Long: `The preview command allows you to quickly view articles
from specific URLs or a list of URLs in a file. It's perfect for
checking out content without subscribing to a feed or saving
anything to your local cache.`,
	RunE: func(cmd *cobra.Command, args []string) error {
		urls, err := cmd.Flags().GetStringSlice("url")
		if err != nil {
			return err
		}

		for _, url := range urls {
			if err := internal.FetchFeed(url); err != nil {
				return err
			} else {
				fmt.Println(url)
			}
		}
		return nil
	},
}

func init() {
	rootCmd.AddCommand(previewCmd)
	previewCmd.Flags().StringSlice("url", []string{}, "URL of the feed to preview")
}
