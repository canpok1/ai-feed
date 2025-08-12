package cmd

import (
	"errors"
	"fmt"

	"github.com/canpok1/ai-feed/cmd/runner"
	"github.com/canpok1/ai-feed/internal/domain"
	"github.com/canpok1/ai-feed/internal/domain/entity"
	"github.com/canpok1/ai-feed/internal/infra"
	"github.com/canpok1/ai-feed/internal/infra/comment"
	"github.com/canpok1/ai-feed/internal/infra/profile"

	"github.com/spf13/cobra"
)

func makeRecommendCmd(fetchClient domain.FetchClient) *cobra.Command {
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
				loadedInfraProfile, loadProfileErr := profile.NewYamlProfileRepositoryImpl(profilePath).LoadInfraProfile()
				if loadProfileErr != nil {
					return fmt.Errorf("failed to load profile from %s: %w", profilePath, loadProfileErr)
				}
				currentProfile.Merge(loadedInfraProfile)
			}

			// AIConfig と PromptConfig を取得
			var aiConfigEntity *entity.AIConfig
			if currentProfile.AI != nil {
				var err error
				aiConfigEntity, err = currentProfile.AI.ToEntity()
				if err != nil {
					return fmt.Errorf("failed to process AI config: %w", err)
				}
			}

			var promptConfigEntity *entity.PromptConfig
			if currentProfile.Prompt != nil {
				promptConfigEntity = currentProfile.Prompt.ToEntity()
			}

			// Recommender を作成
			recommender := domain.NewRandomRecommender(
				comment.NewCommentGeneratorFactory(),
				aiConfigEntity,
				promptConfigEntity,
			)

			recommendRunner, runnerErr := runner.NewRecommendRunner(fetchClient, recommender, cmd.OutOrStdout(), cmd.ErrOrStderr(), currentProfile.Output, currentProfile.Prompt)
			if runnerErr != nil {
				return fmt.Errorf("failed to create runner: %w", runnerErr)
			}

			params, paramsErr := newRecommendParams(cmd)
			if paramsErr != nil {
				return fmt.Errorf("failed to create params: %w", paramsErr)
			}
			err = recommendRunner.Run(cmd.Context(), params, currentProfile)
			if err != nil {
				// 記事が見つからない場合は友好的なメッセージを表示してエラーではない扱いにする
				if errors.Is(err, runner.ErrNoArticlesFound) {
					fmt.Fprintln(cmd.OutOrStdout(), "記事が見つかりませんでした。")
					return nil
				}
				return err
			}
			return nil
		},
	}

	cmd.Flags().StringP("url", "u", "", "URL of the feed to recommend from")
	cmd.Flags().StringP("source", "s", "", "Path to a file containing a list of URLs")
	cmd.Flags().StringP("profile", "p", "", "Path to a profile YAML file")

	cmd.SilenceUsage = true
	return cmd
}

func newRecommendParams(cmd *cobra.Command) (*runner.RecommendParams, error) {
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
		urls, err = infra.ReadURLsFromFile(sourcePath)
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

	return &runner.RecommendParams{
		URLs: urls,
	}, nil
}
