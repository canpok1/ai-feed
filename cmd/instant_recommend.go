package cmd

import (
	"fmt"

	"github.com/spf13/cobra"
)

// instantRecommendCmd represents the instant-recommend command
var instantRecommendCmd = &cobra.Command{
	Use:   "instant-recommend",
	Short: "Recommend a random article from a given URL instantly.",
	Long: `This command fetches articles from the specified URL and
recommends one random article from the fetched list.`,
	Run: func(cmd *cobra.Command, args []string) {
		fmt.Println("instant-recommend called")
	},
}

func init() {
	rootCmd.AddCommand(instantRecommendCmd)

	// Here you will define your flags and configuration settings.

	// Cobra supports Persistent Flags which will work for this command
	// and all subcommands, e.g.:
	// instantRecommendCmd.PersistentFlags().String("foo", "", "A help for foo")

	// Cobra supports local flags which will only run when this command
	// is called directly, e.g.:
	// instantRecommendCmd.Flags().BoolP("toggle", "t", false, "Help message for toggle")
}
