package cmd

import (
	"os"

	"github.com/canpok1/ai-feed/internal/domain"
	"github.com/canpok1/ai-feed/internal/domain/entity"
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
		Args:  cobra.ExactArgs(1),
		Run: func(cmd *cobra.Command, args []string) {
			filePath := args[0]

			// ProfileServiceを使用してバリデーション実行
			validator := domain.NewProfileValidator()
			repoFactory := func(path string) domain.ProfileRepository {
				return &profileRepositoryAdapter{
					repo: infra.NewYamlProfileRepository(path),
				}
			}
			profileService := domain.NewProfileService(validator, repoFactory)
			result, err := profileService.ValidateProfile(filePath)
			if err != nil {
				cmd.PrintErrf("Error: %v\n", err)
				os.Exit(1)
				return
			}

			// 結果の表示
			if !result.IsValid {
				cmd.PrintErrln("Profile validation failed:")
				for _, err := range result.Errors {
					cmd.PrintErrf("  ERROR: %s\n", err)
				}
				os.Exit(2)
				return
			}

			if len(result.Warnings) > 0 {
				cmd.PrintErrln("Profile validation completed with warnings:")
				for _, warning := range result.Warnings {
					cmd.PrintErrf("  WARNING: %s\n", warning)
				}
			} else {
				cmd.Println("Profile validation successful")
			}
		},
	}
	return cmd
}

// profileRepositoryAdapter はinfra.YamlProfileRepositoryをdomain.ProfileRepositoryに適応させるアダプター
type profileRepositoryAdapter struct {
	repo *infra.YamlProfileRepository
}

// LoadProfile はinfraのProfileをentityのProfileに変換して返す
func (a *profileRepositoryAdapter) LoadProfile() (*entity.Profile, error) {
	profile, err := a.repo.LoadProfile()
	if err != nil {
		return nil, err
	}
	return profile.ToEntity(), nil
}

