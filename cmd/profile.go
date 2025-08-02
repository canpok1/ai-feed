package cmd

import (
	"github.com/canpok1/ai-feed/internal/infra"
	"github.com/spf13/cobra"
)

func makeProfileCmd() *cobra.Command {
	cmd := &cobra.Command{
		Use:   "profile",
		Short: "Manage user profiles.",
	}
	profileInitCmd := makeProfileInitCmd()
	cmd.AddCommand(profileInitCmd)
	return cmd
}

func makeProfileInitCmd() *cobra.Command {
	cmd := &cobra.Command{
		Use:   "init [file path]",
		Short: "Initialize a new profile file.",
		Args:  cobra.ExactArgs(1),
		Run: func(cmd *cobra.Command, args []string) {
			filePath := args[0]
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
