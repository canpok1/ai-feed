package cmd

import (
	"fmt"
	"log"
	"math/rand"
	"time"

	"github.com/canpok1/ai-feed/internal"

	"github.com/spf13/cobra"
)

// instantRecommendCmd represents the instant-recommend command
var instantRecommendCmd = &cobra.Command{
	Use:   "instant-recommend",
	Short: "Recommend a random article from a given URL instantly.",
	Long: `This command fetches articles from the specified URL and
recommends one random article from the fetched list.`,
	Run: func(cmd *cobra.Command, args []string) {
		url, _ := cmd.Flags().GetString("url")
		articles, err := internal.FetchFeed(url)
		if err != nil {
			log.Fatalf("Failed to fetch feed: %v", err)
		}

		if len(articles) == 0 {
			fmt.Println("No articles found in the feed.")
			return
		}

		rand.Seed(time.Now().UnixNano())
		randomArticle := articles[rand.Intn(len(articles))]

		fmt.Printf("Title: %s\n", randomArticle.Title)
		fmt.Printf("Link: %s\n", randomArticle.Link)
	},
}

func init() {
	rootCmd.AddCommand(instantRecommendCmd)

	instantRecommendCmd.Flags().StringP("url", "u", "", "URL of the feed to recommend from")
	instantRecommendCmd.MarkFlagRequired("url")
}

