package cmd

import (
	"fmt"
	"os"

	"github.com/canpok1/ai-feed/internal/domain"
	"github.com/canpok1/ai-feed/internal/infra"
	"github.com/spf13/cobra"
)

func makeProfileCmd() *cobra.Command {
	cmd := &cobra.Command{
		Use:   "profile",
		Short: "Manage user profiles.",
	}
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
		Run: func(cmd *cobra.Command, args []string) {
			filePath := args[0]

			// Check if file already exists to avoid accidental overwrites.
			if _, err := os.Stat(filePath); !os.IsNotExist(err) {
				if err == nil {
					cmd.PrintErrf("Error: file already exists at %s. Please specify a new file path.\n", filePath)
				} else {
					cmd.PrintErrf("Error checking file path: %v\n", err)
				}
				return
			}

			profile := infra.NewDefaultProfile()
			profileRepo := infra.NewYamlProfileRepository(filePath)
			err := profileRepo.SaveProfile(profile)
			if err != nil {
				cmd.PrintErrf("Failed to create profile file: %v\n", err)
				return
			}
			cmd.Printf("Profile file created successfully at %s\n", filePath)
		},
	}
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
			config, _ := infra.NewYamlConfigRepository(configPath).Load()
			// 読み込みエラーは無視して処理を継続

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
				loadedProfile, err := infra.NewYamlProfileRepository(filePath).LoadProfile()
				if err != nil {
					cmd.PrintErrf("Error: failed to load profile: %v\n", err)
					return err
				}

				// デフォルトプロファイルとマージ
				currentProfile.Merge(loadedProfile)
			}

			// マージ後のプロファイルをentity.Profileに変換してバリデーション
			entityProfile := currentProfile.ToEntity()
			validator := domain.NewProfileValidator()
			result := validator.Validate(entityProfile)

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
	return cmd
}
