package cmd

import (
	"fmt"
	"io"
	"math/rand"
	"time"

	"github.com/canpok1/ai-feed/internal"

	"github.com/spf13/cobra"
)

func displayArticle(w io.Writer, article internal.Article) {
	fmt.Fprintf(w, "Title: %s\n", article.Title)
	fmt.Fprintf(w, "Link: %s\n", article.Link)
}

// instantRecommendCmd represents the instant-recommend command
var instantRecommendCmd = &cobra.Command{
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

		var urls []string
		if sourcePath != "" {
			urls, err = internal.ReadURLsFromFile(sourcePath)
			if err != nil {
				return fmt.Errorf("failed to read URLs from file: %w", err)
			}
			if len(urls) == 0 {
				return fmt.Errorf("source file contains no URLs")
			}
		} else {
			urls = []string{url}
		}

		var allArticles []internal.Article
		for _, u := range urls {
			articles, err := internal.FetchFeed(u)
			if err != nil {
				fmt.Fprintf(cmd.ErrOrStderr(), "Error fetching feed from %s: %v\n", u, err)
				continue
			}
			allArticles = append(allArticles, articles...)
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

func init() {
	rootCmd.AddCommand(instantRecommendCmd)

	instantRecommendCmd.Flags().StringP("url", "u", "", "URL of the feed to recommend from")
	instantRecommendCmd.Flags().StringP("source", "s", "", "Path to a file containing a list of URLs")
	instantRecommendCmd.MarkFlagRequired("url")
}
