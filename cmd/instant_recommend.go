package cmd

import (
	"fmt"

	"github.com/canpok1/ai-feed/internal"
	"github.com/canpok1/ai-feed/internal/domain"
	"github.com/canpok1/ai-feed/internal/domain/entity"

	"github.com/spf13/cobra"
)

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

			// TODO 設定ファイル読み込み
			config := entity.MakeDefaultConfig()
			model, err := config.GetDefaultAIModel()
			if err != nil {
				return err
			}
			prompt, err := config.GetDefaultPrompt()
			if err != nil {
				return err
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

			recommend, err := recommender.Recommend(
				*model,
				*prompt,
				allArticles)
			if err != nil {
				return fmt.Errorf("failed to recommend article: %w", err)
			}

			viewer, err := domain.NewStdViewer(cmd.OutOrStdout())
			if err != nil {
				return fmt.Errorf("failed to create viewer: %w", err)
			}

			err = viewer.ViewRecommend(recommend)
			if err != nil {
				return fmt.Errorf("failed to view recommend: %w", err)
			}

			return nil
		},
	}

	cmd.Flags().StringP("url", "u", "", "URL of the feed to recommend from")
	cmd.Flags().StringP("source", "s", "", "Path to a file containing a list of URLs")

	return cmd
}
