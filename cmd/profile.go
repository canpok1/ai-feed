package cmd

import (
	"fmt"
	"os"
	"os/user"
	"path/filepath"
	"strings"

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
		Args:  cobra.ExactArgs(1),
		Run: func(cmd *cobra.Command, args []string) {
			filePath := args[0]

			// パス解決を実行
			resolvedPath, err := resolvePath(filePath)
			if err != nil {
				cmd.PrintErrf("Error resolving path: %v\n", err)
				os.Exit(1)
				return
			}

			// ファイルの存在確認
			if _, err := os.Stat(resolvedPath); os.IsNotExist(err) {
				cmd.PrintErrf("Error: profile file not found at %s\n", resolvedPath)
				os.Exit(1)
				return
			} else if err != nil {
				cmd.PrintErrf("Error accessing file: %v\n", err)
				os.Exit(1)
				return
			}

			// プロファイルファイルの読み込み
			profileRepo := infra.NewYamlProfileRepository(resolvedPath)
			profile, err := profileRepo.LoadProfile()
			if err != nil {
				cmd.PrintErrf("Error loading profile: %v\n", err)
				os.Exit(1)
				return
			}

			// バリデーション実行
			validator := domain.NewProfileValidator()
			result := validator.Validate(profile.ToEntity())

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

// resolvePath はパス文字列を解決する（ホームディレクトリや環境変数を展開）
func resolvePath(path string) (string, error) {
	// 環境変数の展開
	path = os.ExpandEnv(path)

	// ホームディレクトリの展開
	if strings.HasPrefix(path, "~/") {
		usr, err := user.Current()
		if err != nil {
			return "", fmt.Errorf("failed to get current user: %w", err)
		}
		path = filepath.Join(usr.HomeDir, path[2:])
	}

	// 絶対パスに変換
	absPath, err := filepath.Abs(path)
	if err != nil {
		return "", fmt.Errorf("failed to resolve absolute path: %w", err)
	}

	return absPath, nil
}

