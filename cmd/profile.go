package cmd

import (
	"fmt"

	"github.com/canpok1/ai-feed/cmd/runner"
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
			r := runner.NewProfileInitRunner(profileRepo, cmd.ErrOrStderr())
			err := r.Run()
			if err != nil {
				return err
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
			r := runner.NewProfileCheckRunner(configPath, cmd.ErrOrStderr(), profileRepoFn)
			result, err := r.Run(profilePath)
			if err != nil {
				return err
			}

			// 結果の表示
			if !result.IsValid {
				cmd.PrintErrln("プロファイルの検証に失敗しました:")
				for _, errMsg := range result.Errors {
					cmd.PrintErrf("  エラー: %s\n", errMsg)
				}
				return fmt.Errorf("プロファイルの検証に失敗しました")
			}

			if len(result.Warnings) > 0 {
				cmd.PrintErrln("プロファイルの検証が警告付きで完了しました:")
				for _, warning := range result.Warnings {
					cmd.PrintErrf("  警告: %s\n", warning)
				}
			} else {
				fmt.Fprintln(cmd.OutOrStdout(), "プロファイルの検証が完了しました")
			}

			return nil
		},
	}
	cmd.SilenceUsage = true
	return cmd
}
