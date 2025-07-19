package cmd

import (
	"fmt"
	"io"

	"github.com/canpok1/ai-feed/internal"
	"github.com/canpok1/ai-feed/internal/domain"
	"github.com/canpok1/ai-feed/internal/infra"

	"github.com/spf13/cobra"
)

func displayRecommend(w io.Writer, recommend *domain.Recommend) {
	if recommend == nil {
		fmt.Fprintln(w, "No articles found in the feed.")
		return
	}

	fmt.Fprintf(w, "Title: %s\n", recommend.Article.Title)
	fmt.Fprintf(w, "Link: %s\n", recommend.Article.Link)
}

func makeInstantRecommendCmd(fetchClient domain.FetchClient, recommender domain.Recommender) *cobra.Command {
	cmd := &cobra.Command{
		Use:   "instant-recommend",
		Short: "Recommend a random article from a given URL instantly.",
		Long: `This command fetches articles from the specified URL and
recommends one random article from the fetched list.`,
		RunE: func(cmd *cobra.Command, args []string) error {
			url, err := cmd.Flags().GetString("url")
			if err != nil {
				return fmt.Errorf("failed to get url flag: %w", err)
			}
			sourcePath, err := cmd.Flags().GetString("source")
			if err != nil {
				return fmt.Errorf("failed to get source flag: %w", err)
			}

			if url != "" && sourcePath != "" {
				return fmt.Errorf("cannot use --url and --source options together")
			}

			var urls []string
			if sourcePath != "" {
				urls, err = internal.ReadURLsFromFile(sourcePath)
				if err != nil {
					return fmt.Errorf("failed to read URLs from file: %w", err)
				}
				if len(urls) == 0 {
					return fmt.Errorf("source file contains no URLs")
				}
			} else if url != "" {
				urls = []string{url}
			} else {
				return fmt.Errorf("either --url or --source must be specified")
			}

			fetcher := domain.NewFetcher(
				fetchClient,
				func(url string, err error) error {
					fmt.Fprintf(cmd.ErrOrStderr(), "Error fetching feed from %s: %v\n", url, err)
					return nil
				},
			)
			allArticles, err := fetcher.Fetch(urls, 0)
			if err != nil {
				return fmt.Errorf("failed to fetch articles: %w", err)
			}

			if len(allArticles) == 0 {
				fmt.Fprintln(cmd.OutOrStdout(), "No articles found in the feed.")
				return nil
			}

			recommend, err := recommender.Recommend(allArticles)
			if err != nil {
				return fmt.Errorf("failed to recommend article: %w", err)
			}

			displayRecommend(cmd.OutOrStdout(), recommend)
			return nil
		},
	}

	cmd.Flags().StringP("url", "u", "", "URL of the feed to recommend from")
	cmd.Flags().StringP("source", "s", "", "Path to a file containing a list of URLs")

	return cmd
}

func init() {
	cmd := makeInstantRecommendCmd(infra.NewFetchClient(), domain.NewRandomRecommender())
	rootCmd.AddCommand(cmd)
}
