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
			// This should not happen as the flag is required, but it's good practice to handle.
			return fmt.Errorf("failed to get url flag: %w", err)
		}
		articles, err := internal.FetchFeed(url)
		if err != nil {
			return fmt.Errorf("failed to fetch feed: %w", err)
		}

		if len(articles) == 0 {
			fmt.Fprintln(cmd.OutOrStdout(), "No articles found in the feed.")
			return nil
		}

		// Use a new random source to avoid seeding the global one.
		r := rand.New(rand.NewSource(time.Now().UnixNano()))
		randomArticle := articles[r.Intn(len(articles))]

		displayArticle(cmd.OutOrStdout(), randomArticle)
		return nil
	},
}

func init() {
	rootCmd.AddCommand(instantRecommendCmd)

	instantRecommendCmd.Flags().StringP("url", "u", "", "URL of the feed to recommend from")
	instantRecommendCmd.MarkFlagRequired("url")
}
