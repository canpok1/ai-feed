package cmd

import (
	"fmt"
	"io"
	"math/rand"
	"time"

	"github.com/canpok1/ai-feed/internal"
	"github.com/canpok1/ai-feed/internal/domain"
	"github.com/canpok1/ai-feed/internal/infra"

	"github.com/spf13/cobra"
)

func displayArticle(w io.Writer, article domain.Article) {
	fmt.Fprintf(w, "Title: %s\n", article.Title)
	fmt.Fprintf(w, "Link: %s\n", article.Link)
}

func makeInstantRecommendCmd(fetchClient domain.FetchClient) *cobra.Command {
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

			r := rand.New(rand.NewSource(time.Now().UnixNano()))
			randomArticle := allArticles[r.Intn(len(allArticles))]

			displayArticle(cmd.OutOrStdout(), randomArticle)
			return nil
		},
	}

	cmd.Flags().StringP("url", "u", "", "URL of the feed to recommend from")
	cmd.Flags().StringP("source", "s", "", "Path to a file containing a list of URLs")

	return cmd
}

func init() {
	cmd := makeInstantRecommendCmd(infra.NewFetchClient())
	rootCmd.AddCommand(cmd)
}
