package cmd

import (
	"fmt"

	"github.com/canpok1/ai-feed/internal/domain"
	"github.com/canpok1/ai-feed/internal/infra"
	"github.com/spf13/cobra"
)

func makePreviewCmd() *cobra.Command {
	cmd := &cobra.Command{
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

			sourceFile, err := cmd.Flags().GetString("source")
			if err != nil {
				return err
			}

			if cmd.Flags().Changed("source") && cmd.Flags().Changed("url") {
				return fmt.Errorf("cannot use --source and --url options together")
			}

			if sourceFile != "" {
				fileURLs, err := infra.ReadURLsFromFile(sourceFile)
				if err != nil {
					return fmt.Errorf("failed to read URLs from file %s: %w", sourceFile, err)
				}
				urls = append(urls, fileURLs...)
			}

			urls = deduplicateURLs(urls)

			limit, err := cmd.Flags().GetInt("limit")
			if err != nil {
				return err
			}

			fetcher := domain.NewFetcher(
				infra.NewFetchClient(),
				func(url string, err error) error {
					fmt.Fprintf(cmd.ErrOrStderr(), "Error fetching feed from %s: %v\n", url, err)
					return nil
				},
			)
			allArticles, err := fetcher.Fetch(urls, limit)
			if err != nil {
				return fmt.Errorf("failed to fetch articles: %w", err)
			}

			viewer, err := infra.NewStdViewer(cmd.OutOrStdout())
			if err != nil {
				return fmt.Errorf("failed to create viewer: %w", err)
			}

			err = viewer.SendArticles(allArticles)
			if err != nil {
				return fmt.Errorf("failed to view articles: %w", err)
			}

			return nil
		},
	}

	cmd.Flags().StringSliceP("url", "u", []string{}, "URL of the feed to preview")
	cmd.Flags().StringP("source", "s", "", "Path to a file containing a list of URLs to preview")
	cmd.Flags().IntP("limit", "l", 0, "Maximum number of articles to display")

	return cmd
}

func deduplicateURLs(urls []string) []string {
	uniqueURLs := make(map[string]bool)
	var finalURLs []string
	for _, url := range urls {
		if _, ok := uniqueURLs[url]; !ok {
			uniqueURLs[url] = true
			finalURLs = append(finalURLs, url)
		}
	}
	return finalURLs
}
