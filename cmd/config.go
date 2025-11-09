package cmd

import (
	"fmt"
	"io"
	"log/slog"

	"github.com/canpok1/ai-feed/internal/domain"
	"github.com/canpok1/ai-feed/internal/domain/entity"
	"github.com/canpok1/ai-feed/internal/infra"
	"github.com/canpok1/ai-feed/internal/infra/profile"
	"github.com/spf13/cobra"
)

func makeConfigCmd() *cobra.Command {
	cmd := &cobra.Command{
		Use:   "config",
		Short: "設定ファイルの管理を行います",
		Long:  `設定ファイルの作成や検証など、設定に関する操作を実行します。`,
	}

	cmd.AddCommand(makeConfigCheckCmd())

	return cmd
}

func makeConfigCheckCmd() *cobra.Command {
	var profilePath string
	var verboseFlag bool

	cmd := &cobra.Command{
		Use:   "check",
		Short: "設定ファイルの内容を検証します",
		Long:  `設定ファイルに必須項目が正しく設定されているかを検証します。`,
		RunE: func(cmd *cobra.Command, args []string) error {
			slog.Debug("Starting config check command")

			// --config フラグの値を取得（グローバルフラグ）
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

			// デフォルトプロファイルをentity.Profileに変換
			var currentProfile *infra.Profile
			if config.DefaultProfile != nil {
				currentProfile = config.DefaultProfile
			} else {
				currentProfile = &infra.Profile{}
			}

			// プロファイルをentity.Profileに変換
			entityProfile, err := currentProfile.ToEntity()
			if err != nil {
				return fmt.Errorf("failed to convert profile to entity: %w", err)
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
				entityProfile.Merge(loadedProfile)
			}

			// バリデーションを実行して結果を返す
			return validateAndPrint(cmd, config, entityProfile, verboseFlag)
		},
	}

	cmd.Flags().StringVarP(&profilePath, "profile", "p", "", "プロファイルYAMLファイルのパス")
	cmd.Flags().BoolVarP(&verboseFlag, "verbose", "v", false, "詳細な設定サマリーを表示")
	cmd.SilenceUsage = true

	return cmd
}

// validateAndPrint はバリデーションを実行して結果を出力する
func validateAndPrint(cmd *cobra.Command, config *infra.Config, entityProfile *entity.Profile, verboseFlag bool) error {
	validator := infra.NewConfigValidator(config, entityProfile)
	result, validateErr := validator.Validate()
	if validateErr != nil {
		return fmt.Errorf("failed to validate config: %w", validateErr)
	}

	// バリデーション結果を出力
	printValidationResult(cmd, result, verboseFlag)

	// バリデーション失敗時は終了コード1
	if !result.Valid {
		return fmt.Errorf("設定ファイルのバリデーションに失敗しました")
	}

	slog.Debug("Config check command completed successfully")
	return nil
}

// printValidationResult はバリデーション結果を出力する
func printValidationResult(cmd *cobra.Command, result *domain.ValidationResult, verboseFlag bool) {
	if result.Valid {
		fmt.Fprintln(cmd.OutOrStdout(), "設定に問題ありません。")
		if verboseFlag {
			printSummary(cmd.OutOrStdout(), result.Summary)
		}
	} else {
		fmt.Fprintln(cmd.ErrOrStderr(), "設定に以下の問題があります：")
		for _, err := range result.Errors {
			fmt.Fprintf(cmd.ErrOrStderr(), "- %s: %s\n", err.Field, err.Message)
		}
	}
}

// printSummary は設定のサマリー情報を出力する
func printSummary(out io.Writer, summary domain.ConfigSummary) {
	fmt.Fprintln(out, "")
	fmt.Fprintln(out, "【設定サマリー】")
	printAISummary(out, summary)
	printPromptSummary(out, summary)
	printOutputSummary(out, summary)
	printCacheSummary(out, summary)
}

// printAISummary はAI設定のサマリーを出力する
func printAISummary(out io.Writer, summary domain.ConfigSummary) {
	fmt.Fprintln(out, "AI設定:")
	if summary.GeminiConfigured {
		fmt.Fprintf(out, "  - Gemini API: 設定済み（モデル: %s）\n", summary.GeminiModel)
	} else {
		fmt.Fprintln(out, "  - Gemini API: 未設定")
	}
}

// printPromptSummary はプロンプト設定のサマリーを出力する
func printPromptSummary(out io.Writer, summary domain.ConfigSummary) {
	fmt.Fprintln(out, "プロンプト設定:")
	if summary.SystemPromptConfigured {
		fmt.Fprintln(out, "  - システムプロンプト: 設定済み")
	} else {
		fmt.Fprintln(out, "  - システムプロンプト: 未設定")
	}
	if summary.CommentPromptConfigured {
		fmt.Fprintln(out, "  - コメントプロンプト: 設定済み")
	} else {
		fmt.Fprintln(out, "  - コメントプロンプト: 未設定")
	}
	if summary.FixedMessageConfigured {
		fmt.Fprintln(out, "  - 固定メッセージ: 設定済み")
	} else {
		fmt.Fprintln(out, "  - 固定メッセージ: 未設定")
	}
}

// printOutputSummary は出力設定のサマリーを出力する
func printOutputSummary(out io.Writer, summary domain.ConfigSummary) {
	fmt.Fprintln(out, "出力設定:")
	if summary.SlackConfigured {
		fmt.Fprintln(out, "  - Slack API: 有効")
		fmt.Fprintf(out, "    - チャンネル: %s\n", summary.SlackChannel)
		if summary.SlackMessageTemplateConfigured {
			fmt.Fprintln(out, "    - メッセージテンプレート: 設定済み")
		} else {
			fmt.Fprintln(out, "    - メッセージテンプレート: 未設定")
		}
	} else {
		fmt.Fprintln(out, "  - Slack API: 無効")
	}
	if summary.MisskeyConfigured {
		fmt.Fprintln(out, "  - Misskey: 有効")
		fmt.Fprintf(out, "    - API URL: %s\n", summary.MisskeyAPIURL)
		if summary.MisskeyMessageTemplateConfigured {
			fmt.Fprintln(out, "    - メッセージテンプレート: 設定済み")
		} else {
			fmt.Fprintln(out, "    - メッセージテンプレート: 未設定")
		}
	} else {
		fmt.Fprintln(out, "  - Misskey: 無効")
	}
}

// printCacheSummary はキャッシュ設定のサマリーを出力する
func printCacheSummary(out io.Writer, summary domain.ConfigSummary) {
	fmt.Fprintln(out, "キャッシュ設定:")
	if summary.CacheEnabled {
		fmt.Fprintln(out, "  - キャッシュ: 有効")
		fmt.Fprintf(out, "    - ファイルパス: %s\n", summary.CacheFilePath)
		fmt.Fprintf(out, "    - 最大エントリ数: %d\n", summary.CacheMaxEntries)
		fmt.Fprintf(out, "    - 保持期間: %d日\n", summary.CacheRetentionDays)
	} else {
		fmt.Fprintln(out, "  - キャッシュ: 無効")
	}
}
