package cmd

import (
	"fmt"
	"io"

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

			runner, err := newInstantRecommendRunner(fetchClient, recommender, cmd.OutOrStdout(), cmd.ErrOrStderr())
			if err != nil {
				return fmt.Errorf("failed to create runner: %w", err)
			}

			return runner.Run(cmd, &instantRecommendParams{
				urls:       urls,
				sourcePath: sourcePath,
				config:     config,
			})
		},
	}

	cmd.Flags().StringP("url", "u", "", "URL of the feed to recommend from")
	cmd.Flags().StringP("source", "s", "", "Path to a file containing a list of URLs")

	return cmd
}

type instantRecommendRunner struct {
	fetchClient domain.FetchClient
	recommender domain.Recommender
	fetcher     *domain.Fetcher
	viewer      domain.Viewer
}

func newInstantRecommendRunner(fetchClient domain.FetchClient, recommender domain.Recommender, stdout io.Writer, stderr io.Writer) (*instantRecommendRunner, error) {
	fetcher := domain.NewFetcher(
		fetchClient,
		func(url string, err error) error {
			fmt.Fprintf(stderr, "Error fetching feed from %s: %v\n", url, err)
			return nil
		},
	)
	viewer, err := domain.NewStdViewer(stdout)
	if err != nil {
		return nil, fmt.Errorf("failed to create viewer: %w", err)
	}

	return &instantRecommendRunner{
		fetchClient: fetchClient,
		recommender: recommender,
		fetcher:     fetcher,
		viewer:      viewer,
	}, nil
}

type instantRecommendParams struct {
	urls       []string
	sourcePath string
	config     *entity.Config
}

func (r *instantRecommendRunner) Run(cmd *cobra.Command, p *instantRecommendParams) error {
	model, err := p.config.GetDefaultAIModel()
	if err != nil {
		return err
	}
	prompt, err := p.config.GetDefaultPrompt()
	if err != nil {
		return err
	}

	allArticles, err := r.fetcher.Fetch(p.urls, 0)
	if err != nil {
		return fmt.Errorf("failed to fetch articles: %w", err)
	}

	if len(allArticles) == 0 {
		fmt.Fprintln(cmd.OutOrStdout(), "No articles found in the feed.")
		return nil
	}

	recommend, err := r.recommender.Recommend(
		*model,
		*prompt,
		allArticles)
	if err != nil {
		return fmt.Errorf("failed to recommend article: %w", err)
	}

	err = r.viewer.ViewRecommend(recommend)
	if err != nil {
		return fmt.Errorf("failed to view recommend: %w", err)
	}

	return nil
}
