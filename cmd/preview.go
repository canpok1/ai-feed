package cmd

import (
	"bufio"
	"fmt"
	"net/url"
	"os"
	"sort"
	"strings"
	"time"

	"github.com/canpok1/ai-feed/internal"
	"github.com/spf13/cobra"
)

var previewCmd = &cobra.Command{
	Use:	"preview",
	Short:	"The preview command temporarily fetches and displays articles from specified URLs or files without subscribing or caching them.",
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

		if sourceFile != "" && cmd.Flags().Changed("source") && cmd.Flags().Changed("url") {
			return fmt.Errorf("cannot use --source and --url options together")
		}

		if sourceFile != "" {
			// Read URLs from file
			fileURLs, err := readURLsFromFile(sourceFile, cmd)
			if err != nil {
				return fmt.Errorf("failed to read URLs from file %s: %w", sourceFile, err)
			}
			urls = append(urls, fileURLs...)
		}

		// Remove duplicate URLs
		uniqueURLs := make(map[string]bool)
		var finalURLs []string
		for _, url := range urls {
			if _, ok := uniqueURLs[url]; !ok {
				uniqueURLs[url] = true
				finalURLs = append(finalURLs, url)
			}
		}
		urls = finalURLs

		limit, err := cmd.Flags().GetInt("limit")
		if err != nil {
			return err
		}

		loc, err := time.LoadLocation("Asia/Tokyo")
		if err != nil {
			return err
		}

		var allArticles []internal.Article
		for _, url := range urls {
			articles, err := internal.FetchFeed(url)
			if err != nil {
				fmt.Fprintf(cmd.ErrOrStderr(), "Error fetching feed from %s: %v\n", url, err)
				continue
			}
			allArticles = append(allArticles, articles...)
		}

		// Sort all articles by published date in descending order
		sort.Slice(allArticles, func(i, j int) bool {
			// Treat articles without a published date as the oldest.
			if allArticles[i].Published == nil {
				return false
			}
			if allArticles[j].Published == nil {
				return true
			}
			return allArticles[i].Published.After(*allArticles[j].Published)
		})

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
	previewCmd.Flags().StringP("source", "s", "", "Path to a file containing a list of URLs to preview")
	previewCmd.Flags().IntP("limit", "l", 0, "Maximum number of articles to display")
}

func readURLsFromFile(filePath string, cmd *cobra.Command) ([]string, error) {
	file, err := os.Open(filePath)
	if err != nil {
		return nil, err
	}
	defer file.Close()

	var urls []string
	scanner := bufio.NewScanner(file)
	for scanner.Scan() {
		line := strings.TrimSpace(scanner.Text())
		if line == "" {
			continue
		}

		// URLバリデーション
		_, err := url.ParseRequestURI(line)
		if err != nil {
			fmt.Fprintf(cmd.ErrOrStderr(), "Warning: Invalid URL in %s: %s (Error: %v)\n", filePath, line, err)
			continue
		}

		urls = append(urls, line)
	}

	if err := scanner.Err(); err != nil {
		return nil, err
	}

	return urls, nil
}

