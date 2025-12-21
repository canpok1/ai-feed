package cmd

import (
	"errors"
	"fmt"
	"log/slog"

	"github.com/canpok1/ai-feed/internal/app"
	"github.com/canpok1/ai-feed/internal/domain"
	"github.com/canpok1/ai-feed/internal/domain/entity"
	"github.com/canpok1/ai-feed/internal/infra"
	"github.com/canpok1/ai-feed/internal/infra/cache"
	"github.com/canpok1/ai-feed/internal/infra/comment"
	"github.com/canpok1/ai-feed/internal/infra/message"
	"github.com/canpok1/ai-feed/internal/infra/profile"
	"github.com/canpok1/ai-feed/internal/infra/selector"
	"github.com/slack-go/slack"

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

			// デフォルトプロファイルを取得
			var currentProfile *entity.Profile
			if config.DefaultProfile != nil {
				currentProfile = config.DefaultProfile
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

			// プロファイルのバリデーション
			validationResult := currentProfile.Validate()
			if !validationResult.IsValid {
				fmt.Fprintln(cmd.ErrOrStderr(), "設定の検証に失敗しました:")
				for _, errMsg := range validationResult.Errors {
					fmt.Fprintf(cmd.ErrOrStderr(), "  エラー: %s\n", errMsg)
				}
				slog.Error("Profile validation failed", "errors", validationResult.Errors)
				return fmt.Errorf("プロファイルの検証に失敗しました")
			}

			// ArticleSelector を作成
			selectorFactory := selector.NewArticleSelectorFactory()
			articleSelector, err := selectorFactory.MakeArticleSelector(currentProfile.AI, currentProfile.Prompt)
			if err != nil {
				return fmt.Errorf("failed to create article selector: %w", err)
			}

			// Recommender を作成
			recommender := domain.NewSelectorBasedRecommender(
				articleSelector,
				comment.NewCommentGeneratorFactory(),
				currentProfile.AI,
				currentProfile.Prompt,
			)

			// キャッシュ設定の取得
			cacheEntity := config.Cache

			// MessageSenderファクトリ関数（インフラ層の実装をラップ）
			senderFactory := func(outputConfig *entity.OutputConfig) ([]domain.MessageSender, error) {
				return createMessageSenders(outputConfig)
			}

			// RecommendCacheファクトリ関数（インフラ層の実装をラップ）
			cacheFactory := func(cacheConfig *entity.CacheConfig) (domain.RecommendCache, error) {
				return createRecommendCache(cacheConfig)
			}

			recommendRunner, runnerErr := app.NewRecommendRunner(
				fetchClient,
				recommender,
				cmd.ErrOrStderr(),
				cmd.OutOrStdout(),
				currentProfile.Output,
				currentProfile.Prompt,
				cacheEntity,
				senderFactory,
				cacheFactory,
			)
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
				if errors.Is(err, app.ErrNoArticlesFound) {
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

func newRecommendParams(cmd *cobra.Command) (*app.RecommendParams, error) {
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

	return &app.RecommendParams{
		URLs: urls,
	}, nil
}

// createMessageSenders はOutputConfigに基づいてMessageSenderのリストを作成する
func createMessageSenders(outputConfig *entity.OutputConfig) ([]domain.MessageSender, error) {
	var senders []domain.MessageSender

	if outputConfig == nil {
		return senders, nil
	}

	if outputConfig.SlackAPI != nil {
		slackConfig := outputConfig.SlackAPI
		if slackConfig.Enabled == nil || !*slackConfig.Enabled {
			slog.Info("Slack API output is disabled (enabled: false)")
		} else {
			// Slackクライアントのオプションを設定
			options := []slack.Option{}
			if slackConfig.APIURL != nil && *slackConfig.APIURL != "" {
				// テスト用：カスタムAPIエンドポイントを設定
				options = append(options, slack.OptionAPIURL(*slackConfig.APIURL))
			}
			slackClient := slack.New(slackConfig.APIToken.Value(), options...)
			slackSender := message.NewSlackSender(slackConfig, slackClient)
			senders = append(senders, slackSender)
		}
	}

	if outputConfig.Misskey != nil {
		misskeyConfig := outputConfig.Misskey
		if misskeyConfig.Enabled == nil || !*misskeyConfig.Enabled {
			slog.Info("Misskey output is disabled (enabled: false)")
		} else {
			misskeySender, senderErr := message.NewMisskeySender(misskeyConfig.APIURL, misskeyConfig.APIToken.Value(), misskeyConfig.MessageTemplate)
			if senderErr != nil {
				return nil, fmt.Errorf("failed to create Misskey sender: %w", senderErr)
			}
			senders = append(senders, misskeySender)
		}
	}

	return senders, nil
}

// createRecommendCache はCacheConfigに基づいてRecommendCacheを作成する
func createRecommendCache(cacheConfig *entity.CacheConfig) (domain.RecommendCache, error) {
	if cacheConfig == nil || cacheConfig.Enabled == nil || !*cacheConfig.Enabled {
		return cache.NewNopCache(), nil
	}

	return cache.NewFileRecommendCache(cacheConfig), nil
}
