package cmd

import (
	"fmt"
	"os"

	"github.com/canpok1/ai-feed/internal/infra"
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

			profileRepo := profile.NewYamlProfileRepositoryImpl(filePath)
			// テンプレートを使用してコメント付きprofile.ymlを生成
			err := profileRepo.SaveProfileWithTemplate()
			if err != nil {
				return fmt.Errorf("failed to create profile file: %w", err)
			}
			cmd.Printf("プロファイルファイルが正常に作成されました: %s\n", filePath)
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
			// config.ymlの読み込み
			configPath := "./config.yml"
			config, err := infra.NewYamlConfigRepository(configPath).Load()
			if err != nil {
				// ファイルが存在しない場合は警告を表示しない
				if _, statErr := os.Stat(configPath); os.IsNotExist(statErr) {
					// ファイルが存在しない場合は何もしない
				} else {
					// ファイルが存在するが読み込み・パースに失敗した場合は警告を表示
					cmd.PrintErrf("警告: %s の読み込みまたは解析に失敗しました。空のデフォルトプロファイルで続行します。エラー: %v\n", configPath, err)
				}
			}

			// デフォルトプロファイルの取得
			var currentProfile infra.Profile
			if config != nil && config.DefaultProfile != nil {
				currentProfile = *config.DefaultProfile
			}
			// 存在しない場合は空のプロファイルを使用（ゼロ値）

			// 引数が指定されている場合は指定ファイルとマージ
			if len(args) > 0 {
				filePath := args[0]

				// ファイルの存在確認
				if _, err := os.Stat(filePath); os.IsNotExist(err) {
					cmd.PrintErrf("エラー: プロファイルファイルが見つかりません: %s\n", filePath)
					return err
				} else if err != nil {
					cmd.PrintErrf("エラー: ファイルへのアクセスに失敗しました: %v\n", err)
					return err
				}

				// 指定されたプロファイルファイルの読み込み
				loadedInfraProfile, err := profile.NewYamlProfileRepositoryImpl(filePath).LoadInfraProfile()
				if err != nil {
					cmd.PrintErrf("エラー: プロファイルの読み込みに失敗しました: %v\n", err)
					return err
				}

				// config.ymlが存在する場合はマージ、存在しない場合はloadedInfraProfileをそのまま使用
				if config != nil && config.DefaultProfile != nil {
					// デフォルトプロファイルとマージ
					currentProfile.Merge(loadedInfraProfile)
				} else {
					// config.ymlが存在しない場合は、読み込んだプロファイルをそのまま使用
					currentProfile = *loadedInfraProfile
				}
			}

			// マージ後のプロファイルをentity.Profileに変換してバリデーション
			entityProfile, err := currentProfile.ToEntity()
			if err != nil {
				return fmt.Errorf("failed to process profile: %w", err)
			}
			result := entityProfile.Validate()

			// 結果の表示
			if !result.IsValid {
				cmd.PrintErrln("プロファイルの検証に失敗しました:")
				for _, err := range result.Errors {
					cmd.PrintErrf("  ERROR: %s\n", err)
				}
				return fmt.Errorf("profile validation failed")
			}

			if len(result.Warnings) > 0 {
				cmd.PrintErrln("プロファイルの検証が警告付きで完了しました:")
				for _, warning := range result.Warnings {
					cmd.PrintErrf("  WARNING: %s\n", warning)
				}
			} else {
				cmd.Println("プロファイルの検証が完了しました")
			}

			return nil
		},
	}
	cmd.SilenceUsage = true
	return cmd
}
