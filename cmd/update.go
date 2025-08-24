package cmd

import (
	"bufio"
	"fmt"
	"os"
	"strings"

	"github.com/canpok1/ai-feed/internal/infra"
	goversion "github.com/hashicorp/go-version"
	"github.com/spf13/cobra"
)

// makeUpdateCmd はupdateコマンドを作成する
func makeUpdateCmd() *cobra.Command {
	cmd := &cobra.Command{
		Use:   "update",
		Short: "ai-feedを最新バージョンに更新します",
		Long: `ai-feedを最新バージョンに更新します。
GitHubのリリースから最新の安定版を取得し、現在のバージョンと比較して更新を行います。`,
		RunE: func(cmd *cobra.Command, args []string) error {
			// --checkフラグが指定された場合は更新チェックのみ
			checkOnly, err := cmd.Flags().GetBool("check")
			if err != nil {
				return fmt.Errorf("checkフラグの取得に失敗しました: %w", err)
			}

			// SelfUpdaterを初期化
			updater := infra.NewSelfUpdater("canpok1", "ai-feed")

			// 現在のバージョンを取得
			currentVersion, err := updater.GetCurrentVersion()
			if err != nil {
				fmt.Fprintf(cmd.ErrOrStderr(), "エラー: 現在のバージョンの取得に失敗しました: %v\n", err)
				return fmt.Errorf("現在のバージョンの取得に失敗しました: %w", err)
			}

			// 最新バージョンを取得
			latest, err := updater.GetLatestVersion()
			if err != nil {
				fmt.Fprintf(cmd.ErrOrStderr(), "エラー: 最新バージョンの取得に失敗しました: %v\n", err)
				fmt.Fprintln(cmd.ErrOrStderr(), "ネットワーク接続を確認してください。")
				return fmt.Errorf("最新バージョンの取得に失敗しました: %w", err)
			}

			// バージョン情報を表示
			fmt.Fprintf(cmd.OutOrStdout(), "現在のバージョン: %s\n", currentVersion)
			fmt.Fprintf(cmd.OutOrStdout(), "最新のバージョン: %s\n", latest.Version)

			// バージョン比較
			currentV, err := goversion.NewVersion(currentVersion)
			if err != nil {
				// "dev"などの不正なバージョン文字列の場合は比較をスキップ
				fmt.Fprintf(cmd.OutOrStdout(), "現在のバージョン '%s' は比較できません。更新を試みます。\n", currentVersion)
			} else {
				latestV, err := goversion.NewVersion(latest.Version)
				if err != nil {
					return fmt.Errorf("最新のバージョン文字列の解析に失敗しました: %w", err)
				}

				if currentV.Equal(latestV) {
					fmt.Fprintln(cmd.OutOrStdout(), "既に最新バージョンです。")
					return nil
				}

				if currentV.GreaterThan(latestV) {
					fmt.Fprintf(cmd.OutOrStdout(), "現在のバージョン (%s) は最新バージョン (%s) よりも新しいです。\n", currentV, latestV)
					return nil
				}
			}

			// --checkオプションの場合はここで終了
			if checkOnly {
				fmt.Fprintln(cmd.OutOrStdout(), "更新が利用可能です。")
				return nil
			}

			// ユーザーに更新の確認を求める
			fmt.Fprintf(cmd.OutOrStdout(), "バージョン %s から %s に更新しますか？ (y/N): ", currentVersion, latest.Version)

			reader := bufio.NewReader(os.Stdin)
			response, err := reader.ReadString('\n')
			if err != nil {
				return fmt.Errorf("入力の読み取りに失敗しました: %w", err)
			}

			// 応答を正規化
			response = strings.TrimSpace(strings.ToLower(response))
			if response != "y" && response != "yes" {
				fmt.Fprintln(cmd.OutOrStdout(), "更新をキャンセルしました。")
				return nil
			}

			// 更新実行
			fmt.Fprintln(cmd.OutOrStdout(), "更新を開始しています...")
			err = updater.UpdateBinary(latest)
			if err != nil {
				fmt.Fprintf(cmd.ErrOrStderr(), "エラー: 更新に失敗しました: %v\n", err)
				return fmt.Errorf("更新に失敗しました: %w", err)
			}

			fmt.Fprintln(cmd.OutOrStdout(), "更新が完了しました。")
			fmt.Fprintln(cmd.OutOrStdout(), "新しいバージョンを確認するには 'ai-feed version' を実行してください。")

			return nil
		},
	}

	// --checkフラグを追加
	cmd.Flags().BoolP("check", "c", false, "更新可能かチェックのみ行い、実際の更新は行わない")

	cmd.SilenceUsage = true
	return cmd
}
