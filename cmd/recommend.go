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
				fmt.Fprintf(cmd.ErrOrStderr(), "エラー: 設定ファイルの読み込みに失敗しました: %s\n", configPath)
				fmt.Fprintln(cmd.ErrOrStderr(), "config.ymlの構文を確認してください。ai-feed init で新しい設定ファイルを生成できます。")
				slog.Error("Failed to load config", "config_path", configPath, "error", loadErr)
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
					fmt.Fprintf(cmd.ErrOrStderr(), "エラー: プロファイルファイルの読み込みに失敗しました: %s\n", profilePath)
					fmt.Fprintln(cmd.ErrOrStderr(), "プロファイルファイルの形式を確認してください。")
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

			// キャッシュ設定の取得（Config.Cacheから）
			var cacheEntity *entity.CacheConfig
			if config.Cache != nil {
				var err error
				cacheEntity, err = config.Cache.ToEntity()
				if err != nil {
					return fmt.Errorf("failed to process cache config: %w", err)
				}
			}

			recommendRunner, runnerErr := runner.NewRecommendRunner(fetchClient, recommender, cmd.ErrOrStderr(), cmd.OutOrStdout(), currentProfile.Output, currentProfile.Prompt, cacheEntity)
			if runnerErr != nil {
				return fmt.Errorf("failed to create runner: %w", runnerErr)
			}

			params, paramsErr := newRecommendParams(cmd)
			if paramsErr != nil {
				return fmt.Errorf("failed to create params: %w", paramsErr)
			}
			// ログ出力
			slog.Debug("RecommendRunner.Run parameters",
				"outputConfig", currentProfile.Output,
				"promptConfig", currentProfile.Prompt,
				"cacheConfig", cacheEntity,
				"profile", currentProfile,
			)
			err = recommendRunner.Run(cmd.Context(), params, currentProfile)
			if err != nil {
				// 記事が見つからない場合は友好的なメッセージを表示してエラーではない扱いにする
				if errors.Is(err, runner.ErrNoArticlesFound) {
					fmt.Fprintln(cmd.OutOrStdout(), "記事が見つかりませんでした。")
					fmt.Fprintln(cmd.ErrOrStderr(), "全てのフィードで記事を取得できませんでした。ネットワーク接続を確認してください。")
					return nil
				}
				slog.Error("Command execution failed", "error", err)
				return err
			}
			slog.Debug("Recommend command completed successfully")
			return nil
		},
	}

	cmd.Flags().StringSliceP("url", "u", []string{}, "推薦元となるフィードのURL（複数指定可）")
	cmd.Flags().StringP("source", "s", "", "URLリストを含むファイルのパス")
	cmd.Flags().StringP("profile", "p", "", "プロファイルYAMLファイルのパス")

	cmd.SilenceUsage = true
	return cmd
}

func newRecommendParams(cmd *cobra.Command) (*runner.RecommendParams, error) {
	urlList, err := cmd.Flags().GetStringSlice("url")
	if err != nil {
		return nil, fmt.Errorf("failed to get url flag: %w", err)
	}
	sourcePath, err := cmd.Flags().GetString("source")
	if err != nil {
		return nil, fmt.Errorf("failed to get source flag: %w", err)
	}

	var urls []string

	// --source オプションが指定されている場合、ファイルからURLを読み込む
	if sourcePath != "" {
		sourceURLs, err := infra.ReadURLsFromFile(sourcePath)
		if err != nil {
			return nil, fmt.Errorf("failed to read URLs from file: %w", err)
		}
		urls = append(urls, sourceURLs...)
	}

	// -u オプションで指定されたURLを追加
	if len(urlList) > 0 {
		urls = append(urls, urlList...)
	}

	// いずれのオプションも指定されていない場合はエラー
	if len(urls) == 0 {
		return nil, fmt.Errorf("--url または --source のいずれかを指定してください")
	}

	return &runner.RecommendParams{
		URLs: urls,
	}, nil
}
