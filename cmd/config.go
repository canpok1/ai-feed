package cmd

import (
	"fmt"
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
			fmt.Fprintln(cmd.OutOrStdout(), "")
			fmt.Fprintln(cmd.OutOrStdout(), "【設定サマリー】")
			fmt.Fprintln(cmd.OutOrStdout(), "AI設定:")
			if result.Summary.GeminiConfigured {
				fmt.Fprintf(cmd.OutOrStdout(), "  - Gemini API: 設定済み（モデル: %s）\n", result.Summary.GeminiModel)
			} else {
				fmt.Fprintln(cmd.OutOrStdout(), "  - Gemini API: 未設定")
			}
			fmt.Fprintln(cmd.OutOrStdout(), "プロンプト設定:")
			if result.Summary.CommentPromptConfigured {
				fmt.Fprintln(cmd.OutOrStdout(), "  - コメントプロンプト: 設定済み")
			} else {
				fmt.Fprintln(cmd.OutOrStdout(), "  - コメントプロンプト: 未設定")
			}
			fmt.Fprintln(cmd.OutOrStdout(), "出力設定:")
			if result.Summary.SlackConfigured {
				fmt.Fprintln(cmd.OutOrStdout(), "  - Slack API: 設定済み")
			} else {
				fmt.Fprintln(cmd.OutOrStdout(), "  - Slack API: 未設定")
			}
			if result.Summary.MisskeyConfigured {
				fmt.Fprintln(cmd.OutOrStdout(), "  - Misskey: 設定済み")
			} else {
				fmt.Fprintln(cmd.OutOrStdout(), "  - Misskey: 未設定")
			}
		}
	} else {
		fmt.Fprintln(cmd.ErrOrStderr(), "設定に以下の問題があります：")
		for _, err := range result.Errors {
			fmt.Fprintf(cmd.ErrOrStderr(), "- %s: %s\n", err.Field, err.Message)
		}
	}
}
