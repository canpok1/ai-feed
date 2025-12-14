package cmd

import (
	"fmt"

	"github.com/canpok1/ai-feed/internal/app"
	"github.com/canpok1/ai-feed/internal/domain"
	"github.com/canpok1/ai-feed/internal/infra/profile"
	"github.com/spf13/cobra"
)

func makeProfileCmd() *cobra.Command {
	cmd := &cobra.Command{
		Use:   "profile",
		Short: "ユーザープロファイルを管理します",
	}
	cmd.SilenceUsage = true
	profileInitCmd := makeProfileInitCmd()
	profileCheckCmd := makeProfileCheckCmd()
	cmd.AddCommand(profileInitCmd)
	cmd.AddCommand(profileCheckCmd)
	return cmd
}

func makeProfileInitCmd() *cobra.Command {
	cmd := &cobra.Command{
		Use:   "init [file path]",
		Short: "新しいプロファイルファイルを初期化します",
		Args:  cobra.ExactArgs(1),
		RunE: func(cmd *cobra.Command, args []string) error {
			filePath := args[0]

			// 進行状況メッセージ: 初期化開始
			fmt.Fprintf(cmd.ErrOrStderr(), "プロファイルを初期化しています... (%s)\n", filePath)

			profileRepo := profile.NewYamlProfileRepositoryImpl(filePath)
			r, err := app.NewProfileInitRunner(profileRepo, cmd.ErrOrStderr())
			if err != nil {
				return fmt.Errorf("failed to create runner: %w", err)
			}
			if runErr := r.Run(); runErr != nil {
				return runErr
			}
			// 完了メッセージ（stdout）
			cmd.Printf("プロファイルファイルを作成しました: %s\n", filePath)
			return nil
		},
	}
	cmd.SilenceUsage = true
	return cmd
}

// makeProfileCheckCmd はプロファイルファイルの検証を行うコマンドを作成する
func makeProfileCheckCmd() *cobra.Command {
	cmd := &cobra.Command{
		Use:   "check [file path]",
		Short: "プロファイルファイルの設定を検証します",
		Args:  cobra.MaximumNArgs(1),
		RunE: func(cmd *cobra.Command, args []string) error {
			// --config フラグの値を取得（グローバルフラグ）
			configPath := cfgFile
			if configPath == "" {
				configPath = "./config.yml"
			}

			// 引数からプロファイルファイルのパスを取得
			profilePath := ""
			if len(args) > 0 {
				profilePath = args[0]
			}

			// 進行状況メッセージ: 検証開始（使用するファイルパスを表示）
			fmt.Fprintf(cmd.ErrOrStderr(), "プロファイルを検証しています...\n")
			fmt.Fprintf(cmd.ErrOrStderr(), "  設定ファイル: %s\n", configPath)
			if profilePath != "" {
				fmt.Fprintf(cmd.ErrOrStderr(), "  プロファイル: %s\n", profilePath)
			}

			// ProfileCheckRunnerを使用して検証を実行
			profileRepoFn := func(path string) domain.ProfileRepository {
				return profile.NewYamlProfileRepositoryImpl(path)
			}
			r := app.NewProfileCheckRunner(configPath, cmd.ErrOrStderr(), profileRepoFn)
			result, err := r.Run(profilePath)
			if err != nil {
				// SilenceErrorsが有効なので、手動でエラーを出力
				fmt.Fprintf(cmd.ErrOrStderr(), "Error: %v\n", err)
				return err
			}

			// 結果の表示（統一形式: 1行目=処理完了報告、2行目以降=結果報告）
			if !result.IsValid {
				// 失敗時はすべてstderrに出力
				fmt.Fprintln(cmd.ErrOrStderr(), "プロファイルの検証が完了しました")
				fmt.Fprintln(cmd.ErrOrStderr(), "以下の問題があります：")
				for _, errMsg := range result.Errors {
					fmt.Fprintf(cmd.ErrOrStderr(), "- %s\n", errMsg)
				}
				return fmt.Errorf("プロファイルの検証に失敗しました")
			}

			// 成功時はstdoutに出力
			fmt.Fprintln(cmd.OutOrStdout(), "プロファイルの検証が完了しました")
			if len(result.Warnings) > 0 {
				fmt.Fprintln(cmd.ErrOrStderr(), "以下の警告があります：")
				for _, warning := range result.Warnings {
					fmt.Fprintf(cmd.ErrOrStderr(), "- %s\n", warning)
				}
			} else {
				fmt.Fprintln(cmd.OutOrStdout(), "問題ありませんでした")
			}

			return nil
		},
	}
	cmd.SilenceUsage = true
	cmd.SilenceErrors = true
	return cmd
}
