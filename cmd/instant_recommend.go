package cmd

import (
	"fmt"
	"io"

	"github.com/canpok1/ai-feed/internal"
	"github.com/canpok1/ai-feed/internal/domain"
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
			config, loadErr := infra.NewYamlConfigRepository(configPath).Load()
			if loadErr != nil {
				return fmt.Errorf("failed to load config: %w", loadErr)
			}

			outputConfigs, outputErr := config.GetDefaultOutputs()
			if outputErr != nil {
				return fmt.Errorf("failed to get default outputs: %w", outputErr)
			}

			runner, runnerErr := newInstantRecommendRunner(fetchClient, recommender, cmd.OutOrStdout(), cmd.ErrOrStderr(), outputConfigs)
			if runnerErr != nil {
				return fmt.Errorf("failed to create runner: %w", runnerErr)
			}

			params, paramsErr := newInstantRecommendParams(cmd)
			if paramsErr != nil {
				return fmt.Errorf("failed to create params: %w", paramsErr)
			}
			return runner.Run(cmd, params, config)
		},
	}

	cmd.Flags().StringP("url", "u", "", "URL of the feed to recommend from")
	cmd.Flags().StringP("source", "s", "", "Path to a file containing a list of URLs")

	return cmd
}

type instantRecommendParams struct {
	urls []string
}

func newInstantRecommendParams(cmd *cobra.Command) (*instantRecommendParams, error) {
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

	return &instantRecommendParams{
		urls: urls,
	}, nil
}

type instantRecommendRunner struct {
	fetcher     *domain.Fetcher
	recommender domain.Recommender
	viewers     []domain.Viewer
}

func newInstantRecommendRunner(fetchClient domain.FetchClient, recommender domain.Recommender, stdout io.Writer, stderr io.Writer, outputConfigs []*infra.OutputConfig) (*instantRecommendRunner, error) {
	fetcher := domain.NewFetcher(
		fetchClient,
		func(url string, err error) error {
			fmt.Fprintf(stderr, "Error fetching feed from %s: %v\n", url, err)
			return err
		},
	)
	viewer, err := domain.NewStdViewer(stdout)
	if err != nil {
		return nil, fmt.Errorf("failed to create viewer: %w", err)
	}
	viewers := []domain.Viewer{viewer}

	for _, c := range outputConfigs {
		switch c.Type {
		case "slack-api":
			if c.SlackAPIConfig == nil {
				fmt.Fprintf(stderr, "Warning: slack-api output type found but SlackAPIConfig is nil, skipping\n")
				continue
			}
			slackViewer := infra.NewSlackViewer(c.SlackAPIConfig.ToEntity())
			viewers = append(viewers, slackViewer)
		case "misskey":
			if c.MisskeyConfig == nil {
				fmt.Fprintf(stderr, "Warning: misskey output type found but MisskeyConfig is nil, skipping\n")
				continue
			}
			// TODO: MisskeyViewer の実装と初期化
			// misskeyViewer := infra.NewMisskeyViewer(c.MisskeyConfig);
			// viewers = append(viewers, misskeyViewer);
			fmt.Fprintf(stderr, "Warning: misskey output type is not yet supported, skipping\n")
		default:
			fmt.Fprintf(stderr, "Warning: unsupported output type '%s' found, skipping\n", c.Type)
		}
	}

	return &instantRecommendRunner{
		fetcher:     fetcher,
		recommender: recommender,
		viewers:     viewers,
	}, nil
}

func (r *instantRecommendRunner) Run(cmd *cobra.Command, p *instantRecommendParams, configRepo infra.ConfigRepository) error {
	model, err := configRepo.GetDefaultAIModel()
	if err != nil {
		return fmt.Errorf("failed to get default AI model: %w", err)
	}
	prompt, err := configRepo.GetDefaultPrompt()
	if err != nil {
		return fmt.Errorf("failed to get default prompt: %w", err)
	}
	systemPrompt, err := configRepo.GetDefaultSystemPrompt()
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
		model.ToEntity(),
		prompt.ToEntity(),
		systemPrompt,
		allArticles)
	if err != nil {
		return fmt.Errorf("failed to recommend article: %w", err)
	}

	var errs []error
	for _, viewer := range r.viewers {
		err = viewer.ViewRecommend(recommend)
		if err != nil {
			errs = append(errs, fmt.Errorf("failed to view recommend: %w", err))
		}
	}

	if len(errs) > 0 {
		return fmt.Errorf("failed to view all recommends: %v", errs)
	}

	return nil
}
