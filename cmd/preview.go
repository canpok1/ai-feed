package cmd

import (
	"fmt"
	"time"

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

		limit, err := cmd.Flags().GetInt("limit")
		if err != nil {
			return err
		}

		loc, err := time.LoadLocation("Asia/Tokyo")
		if err != nil {
			return err
		}

		var allArticles []*internal.Article
		for _, url := range urls {
			articles, err := internal.FetchFeed(url)
			if err != nil {
				fmt.Fprintf(cmd.ErrOrStderr(), "Error fetching feed from %s: %v\n", url, err)
				continue
			}
			allArticles = append(allArticles, articles...)
		}

		// Apply limit
		if limit > 0 && len(allArticles) > limit {
			allArticles = allArticles[:limit]
		}

		for _, article := range allArticles {
			fmt.Printf("Title: %s\n", article.Title)
			fmt.Printf("Link: %s\n", article.Link)
			if article.Published != nil {
				fmt.Printf("Published: %s\n", article.Published.In(loc).Format("2006-01-02 15:04:05 JST"))
			}
			fmt.Printf("Content: %s\n", article.Content)
			fmt.Println("---")
		}
		return nil
	},
}

func init() {
	rootCmd.AddCommand(previewCmd)
	previewCmd.Flags().StringSliceP("url", "u", []string{}, "URL of the feed to preview")
	previewCmd.Flags().IntP("limit", "l", 0, "Maximum number of articles to display")
}
