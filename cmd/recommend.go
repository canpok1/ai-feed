package cmd

import (
	"errors"
	"fmt"
	"log/slog"

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
		Short: "指定されたURLからランダムな記事を推薦します",
		Long: `指定されたURLから記事を取得し、その中からランダムに選択した
記事を推薦します。`,
		RunE: func(cmd *cobra.Command, args []string) error {
			slog.Debug("Starting recommend command")
			configPath := cfgFile
			if configPath == "" {
				configPath = "./config.yml"
			}
			slog.Debug("Loading config", "config_path", configPath)
			config, loadErr := infra.NewYamlConfigRepository(configPath).Load()
			if loadErr != nil {
				slog.Error("Failed to load config", "error", loadErr)
				return fmt.Errorf("failed to load config: %w", loadErr)
			}

			profilePath, err := cmd.Flags().GetString("profile")
			if err != nil {
				return fmt.Errorf("failed to get profile flag: %w", err)
			}

			// デフォルトプロファイルをentity.Profileに変換
			var currentProfile *entity.Profile
			if config.DefaultProfile != nil {
				var err error
				currentProfile, err = config.DefaultProfile.ToEntity()
				if err != nil {
					return fmt.Errorf("failed to process default profile: %w", err)
				}
			} else {
				currentProfile = &entity.Profile{}
			}

			// プロファイルファイルが指定されている場合は読み込んでマージ
			if profilePath != "" {
				slog.Debug("Loading profile", "profile_path", profilePath)
				loadedProfile, loadProfileErr := profile.NewYamlProfileRepositoryImpl(profilePath).LoadProfile()
				if loadProfileErr != nil {
					slog.Error("Failed to load profile", "profile_path", profilePath, "error", loadProfileErr)
					return fmt.Errorf("failed to load profile from %s: %w", profilePath, loadProfileErr)
				}
				currentProfile.Merge(loadedProfile)
			}

			// Recommender を作成
			recommender := domain.NewRandomRecommender(
				comment.NewCommentGeneratorFactory(),
				currentProfile.AI,
				currentProfile.Prompt,
			)

			// entity.ProfileからinfraのProfileを構築する必要がある（一時的な処理）
			infraProfile := &infra.Profile{}
			if currentProfile.Output != nil {
				// entity.OutputConfig -> infra.OutputConfigの変換は複雑なため、一時的にnilを渡す
			}
			recommendRunner, runnerErr := runner.NewRecommendRunner(fetchClient, recommender, cmd.ErrOrStderr(), nil, nil)
			if runnerErr != nil {
				return fmt.Errorf("failed to create runner: %w", runnerErr)
			}

			params, paramsErr := newRecommendParams(cmd)
			if paramsErr != nil {
				return fmt.Errorf("failed to create params: %w", paramsErr)
			}
			err = recommendRunner.Run(cmd.Context(), params, *infraProfile)
			if err != nil {
				// 記事が見つからない場合は友好的なメッセージを表示してエラーではない扱いにする
				if errors.Is(err, runner.ErrNoArticlesFound) {
					fmt.Fprintln(cmd.OutOrStdout(), "記事が見つかりませんでした。")
					return nil
				}
				slog.Error("Command execution failed", "error", err)
				return err
			}
			slog.Debug("Recommend command completed successfully")
			return nil
		},
	}

	cmd.Flags().StringP("url", "u", "", "推薦元となるフィードのURL")
	cmd.Flags().StringP("source", "s", "", "URLリストを含むファイルのパス")
	cmd.Flags().StringP("profile", "p", "", "プロファイルYAMLファイルのパス")

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
		return nil, fmt.Errorf("--url と --source オプションは同時に使用できません")
	}

	var urls []string
	if sourcePath != "" {
		urls, err = infra.ReadURLsFromFile(sourcePath)
		if err != nil {
			return nil, fmt.Errorf("failed to read URLs from file: %w", err)
		}
		if len(urls) == 0 {
			return nil, fmt.Errorf("ソースファイルにURLが含まれていません")
		}
	} else if url != "" {
		urls = []string{url}
	} else {
		return nil, fmt.Errorf("--url または --source のいずれかを指定してください")
	}

	return &runner.RecommendParams{
		URLs: urls,
	}, nil
}
