package cmd

import (
	"fmt"
	"io"

	"github.com/canpok1/ai-feed/internal"
	"github.com/canpok1/ai-feed/internal/domain"
	"github.com/canpok1/ai-feed/internal/infra"

	"github.com/spf13/cobra"
)

func makeRecommendCmd(fetchClient domain.FetchClient, recommender domain.Recommender) *cobra.Command {
	cmd := &cobra.Command{
		Use:   "recommend",
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

			profilePath, err := cmd.Flags().GetString("profile")
			if err != nil {
				return fmt.Errorf("failed to get profile flag: %w", err)
			}

			var currentProfile infra.Profile
			if config.DefaultProfile != nil {
				currentProfile = *config.DefaultProfile
			}

			if profilePath != "" {
				loadedProfile, loadProfileErr := infra.NewYamlProfileRepository(profilePath).LoadProfile()
				if loadProfileErr != nil {
					return fmt.Errorf("failed to load profile from %s: %w", profilePath, loadProfileErr)
				}
				currentProfile.Merge(loadedProfile)
			}

			runner, runnerErr := newRecommendRunner(fetchClient, recommender, cmd.OutOrStdout(), cmd.ErrOrStderr(), currentProfile.Output, currentProfile.Prompt)
			if runnerErr != nil {
				return fmt.Errorf("failed to create runner: %w", runnerErr)
			}

			params, paramsErr := newRecommendParams(cmd)
			if paramsErr != nil {
				return fmt.Errorf("failed to create params: %w", paramsErr)
			}
			return runner.Run(cmd, params, currentProfile)
		},
	}

	cmd.Flags().StringP("url", "u", "", "URL of the feed to recommend from")
	cmd.Flags().StringP("source", "s", "", "Path to a file containing a list of URLs")
	cmd.Flags().StringP("profile", "p", "", "Path to a profile YAML file")

	return cmd
}

type recommendParams struct {
	urls []string
}

func newRecommendParams(cmd *cobra.Command) (*recommendParams, error) {
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

	return &recommendParams{
		urls: urls,
	}, nil
}

type recommendRunner struct {
	fetcher     *domain.Fetcher
	recommender domain.Recommender
	viewers     []domain.Viewer
}

func newRecommendRunner(fetchClient domain.FetchClient, recommender domain.Recommender, stdout io.Writer, stderr io.Writer, outputConfig *infra.OutputConfig, promptConfig *infra.PromptConfig) (*recommendRunner, error) {
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

	if outputConfig != nil {
		if outputConfig.SlackAPI != nil {
			slackViewer := infra.NewSlackViewer(outputConfig.SlackAPI.ToEntity())
			viewers = append(viewers, slackViewer)
		}
		if outputConfig.Misskey != nil {
			misskeyViewer, err := infra.NewMisskeyViewer(outputConfig.Misskey.APIURL, outputConfig.Misskey.APIToken)
			if err != nil {
				return nil, fmt.Errorf("failed to create Misskey viewer: %w", err)
			}
			viewers = append(viewers, misskeyViewer)
		}
	}

	return &recommendRunner{
		fetcher:     fetcher,
		recommender: recommender,
		viewers:     viewers,
	}, nil
}

func (r *recommendRunner) Run(cmd *cobra.Command, p *recommendParams, profile infra.Profile) error {
	allArticles, err := r.fetcher.Fetch(p.urls, 0)
	if err != nil {
		return fmt.Errorf("failed to fetch articles: %w", err)
	}

	if len(allArticles) == 0 {
		fmt.Fprintln(cmd.OutOrStdout(), "No articles found in the feed.")
		return nil
	}

	if profile.AI == nil || profile.Prompt == nil {
		return fmt.Errorf("AI model or prompt is not configured")
	}

	aiConfigEntity := profile.AI.ToEntity()
	promptConfigEntity := profile.Prompt.ToEntity()

	recommend, err := r.recommender.Recommend(
		cmd.Context(),
		aiConfigEntity,
		promptConfigEntity,
		allArticles)
	if err != nil {
		return fmt.Errorf("failed to recommend article: %w", err)
	}

	var errs []error
	for _, viewer := range r.viewers {
		err = viewer.ViewRecommend(recommend, profile.Prompt.FixedMessage)
		if err != nil {
			errs = append(errs, fmt.Errorf("failed to view recommend: %w", err))
		}
	}

	if len(errs) > 0 {
		return fmt.Errorf("failed to view all recommends: %v", errs)
	}

	return nil
}
