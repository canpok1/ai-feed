package cmd

import (
	"github.com/canpok1/ai-feed/cmd/runner"
	"github.com/canpok1/ai-feed/internal/domain"
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
			// --config フラグの値を取得（グローバルフラグ）
			configPath := cfgFile
			if configPath == "" {
				configPath = "./config.yml"
			}

			// ProfileRepositoryのファクトリ関数
			profileRepoFn := func(path string) domain.ProfileRepository {
				return profile.NewYamlProfileRepositoryImpl(path)
			}

			// ConfigCheckRunnerを作成して実行
			configCheckRunner := runner.NewConfigCheckRunner(configPath, cmd.OutOrStdout(), cmd.ErrOrStderr(), profileRepoFn)
			params := &runner.ConfigCheckParams{
				ProfilePath: profilePath,
				VerboseFlag: verboseFlag,
			}

			return configCheckRunner.Run(params)
		},
	}

	cmd.Flags().StringVarP(&profilePath, "profile", "p", "", "プロファイルYAMLファイルのパス")
	cmd.Flags().BoolVarP(&verboseFlag, "verbose", "v", false, "詳細な設定サマリーを表示")
	cmd.SilenceUsage = true
	cmd.SilenceErrors = true

	return cmd
}
