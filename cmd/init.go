package cmd

import (
	"fmt"

	"github.com/canpok1/ai-feed/cmd/runner"
	"github.com/canpok1/ai-feed/internal/infra"
	"github.com/spf13/cobra"
)

func makeInitCmd() *cobra.Command {
	cmd := &cobra.Command{
		Use:   "init",
		Short: "設定ファイル（config.yml）のテンプレートを生成します（既存ファイルは上書きしません）",
		RunE: func(cmd *cobra.Command, args []string) error {
			filePath := runner.DefaultConfigFilePath

			// 進行状況メッセージ: 初期化開始
			fmt.Fprintln(cmd.ErrOrStderr(), "設定ファイルを初期化しています...")

			// 依存性の注入
			configRepo := infra.NewYamlConfigRepository(filePath)

			// ConfigInitRunnerを使用してビジネスロジックを実行
			r := runner.NewConfigInitRunner(configRepo, cmd.ErrOrStderr())
			if err := r.Run(); err != nil {
				return err
			}

			// 完了メッセージ（stdout）
			fmt.Fprintf(cmd.OutOrStdout(), "%s を生成しました\n", filePath)
			return nil
		},
	}
	cmd.SilenceUsage = true
	return cmd
}
