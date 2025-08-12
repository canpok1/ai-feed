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
		Short: "Manage user profiles.",
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
		Short: "Initialize a new profile file.",
		Args:  cobra.ExactArgs(1),
		RunE: func(cmd *cobra.Command, args []string) error {
			filePath := args[0]

			profileRepo := profile.NewYamlProfileRepositoryImpl(filePath)
			// テンプレートを使用してコメント付きprofile.ymlを生成
			err := profileRepo.SaveProfileWithTemplate()
			if err != nil {
				return fmt.Errorf("failed to create profile file: %w", err)
			}
			cmd.Printf("Profile file created successfully at %s\n", filePath)
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
		Short: "Validate profile file configuration.",
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
					cmd.PrintErrf("Warning: failed to load or parse %s, proceeding with empty default profile. Error: %v\n", configPath, err)
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
					cmd.PrintErrf("Error: profile file not found at %s\n", filePath)
					return err
				} else if err != nil {
					cmd.PrintErrf("Error: failed to access file: %v\n", err)
					return err
				}

				// 指定されたプロファイルファイルの読み込み
				loadedInfraProfile, err := profile.NewYamlProfileRepositoryImpl(filePath).LoadInfraProfile()
				if err != nil {
					cmd.PrintErrf("Error: failed to load profile: %v\n", err)
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
				cmd.PrintErrln("Profile validation failed:")
				for _, err := range result.Errors {
					cmd.PrintErrf("  ERROR: %s\n", err)
				}
				return fmt.Errorf("profile validation failed")
			}

			if len(result.Warnings) > 0 {
				cmd.PrintErrln("Profile validation completed with warnings:")
				for _, warning := range result.Warnings {
					cmd.PrintErrf("  WARNING: %s\n", warning)
				}
			} else {
				cmd.Println("Profile validation successful")
			}

			return nil
		},
	}
	cmd.SilenceUsage = true
	return cmd
}
