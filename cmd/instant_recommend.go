package cmd

import (
	"fmt"
	"io"

	"github.com/canpok1/ai-feed/internal"
	"github.com/canpok1/ai-feed/internal/domain"
	"github.com/canpok1/ai-feed/internal/domain/entity"
	"github.com/canpok1/ai-feed/internal/infra"

	"github.com/spf13/cobra"
)

func makeInstantRecommendCmd(fetchClient domain.FetchClient, recommender domain.Recommender) *cobra.Command {
	cmd := &cobra.Command{
		Use:   "instant-recommend",
		Short: "Recommend a random article from a given URL instantly.",
		Long: `This command fetches articles from the specified URL and
recommends one random article from the fetched list.`,
		RunE: func(cmd *cobra.Command, args []string) error {
			configPath := cfgFile
			if configPath == "" {
				configPath = "./config.yml"
			}
			configRepo := infra.NewYamlConfigRepository(configPath)
			params, err := newInstantRecommendParams(cmd, configRepo)
			if err != nil {
				return fmt.Errorf("failed to create params: %w", err)
			}

			outputConfigs := []entity.OutputConfig{}
			for _, outputConfig := range params.config.Outputs {
				outputConfigs = append(outputConfigs, outputConfig)
			}

			runner, err := newInstantRecommendRunner(fetchClient, recommender, cmd.OutOrStdout(), cmd.ErrOrStderr(), outputConfigs)
			if err != nil {
				return fmt.Errorf("failed to create runner: %w", err)
			}

			return runner.Run(cmd, params)
		},
	}

	cmd.Flags().StringP("url", "u", "", "URL of the feed to recommend from")
	cmd.Flags().StringP("source", "s", "", "Path to a file containing a list of URLs")

	return cmd
}

type instantRecommendParams struct {
	urls   []string
	config *entity.Config
}

func newInstantRecommendParams(cmd *cobra.Command, configRepo domain.ConfigRepository) (*instantRecommendParams, error) {
	url, err := cmd.Flags().GetString("url")
	if err != nil {
		return nil, fmt.Errorf("failed to get url flag: %w", err)
	}
	sourcePath, err := cmd.Flags().GetString("source")
	if err != nil {
		return nil, fmt.Errorf("failed to get source flag: %w", err)
	}

	if url != "" && sourcePath != "" {
		return nil, fmt.Errorf("cannot use --url and --source options together")
	}

	var urls []string
	if sourcePath != "" {
		urls, err = internal.ReadURLsFromFile(sourcePath)
		if err != nil {
			return nil, fmt.Errorf("failed to read URLs from file: %w", err)
		}
		if len(urls) == 0 {
			return nil, fmt.Errorf("source file contains no URLs")
		}
	} else if url != "" {
		urls = []string{url}
	} else {
		return nil, fmt.Errorf("either --url or --source must be specified")
	}

	config, err := configRepo.Load()
	if err != nil {
		return nil, fmt.Errorf("failed to load config: %w", err)
	}

	return &instantRecommendParams{
		urls:   urls,
		config: config,
	}, nil
}

type instantRecommendRunner struct {
	fetcher     *domain.Fetcher
	recommender domain.Recommender
	viewers     []domain.Viewer
}

func newInstantRecommendRunner(fetchClient domain.FetchClient, recommender domain.Recommender, stdout io.Writer, stderr io.Writer, outputConfigs []entity.OutputConfig) (*instantRecommendRunner, error) {
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
	viewers := []domain.Viewer{viewer}

	for _, c := range outputConfigs {
		if c.Type == "slack-api" {
			slackViewer := infra.NewSlackViewer(c.APIToken, c.Channel)
			viewers = append(viewers, slackViewer)
		} else {
			return nil, fmt.Errorf("unsupported output type: %s", c.Type)
		}
	}

	return &instantRecommendRunner{
		fetcher:     fetcher,
		recommender: recommender,
		viewers:     viewers,
	}, nil
}

func (r *instantRecommendRunner) Run(cmd *cobra.Command, p *instantRecommendParams) error {
	model, err := p.config.GetDefaultAIModel()
	if err != nil {
		return fmt.Errorf("failed to get default AI model: %w", err)
	}
	prompt, err := p.config.GetDefaultPrompt()
	if err != nil {
		return fmt.Errorf("failed to get default prompt: %w", err)
	}
	systemPrompt, err := p.config.GetDefaultSystemPrompt()
	if err != nil {
		return fmt.Errorf("failed to get default system prompt: %w", err)
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
		cmd.Context(),
		model,
		prompt,
		systemPrompt,
		allArticles)
	if err != nil {
		return fmt.Errorf("failed to recommend article: %w", err)
	}

	for _, viewer := range r.viewers {
		err = viewer.ViewRecommend(recommend)
		if err != nil {
			return fmt.Errorf("failed to view recommend: %w", err)
		}
	}

	return nil
}
