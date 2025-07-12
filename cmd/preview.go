package cmd

import (
	"fmt"

	"github.com/spf13/cobra"
)

var previewCmd = &cobra.Command{
	Use:   "preview",
	Short: "The preview command temporarily fetches and displays articles from specified URLs or files without subscribing or caching them.",
	Long: `The preview command allows you to quickly view articles
from specific URLs or a list of URLs in a file. It's perfect for
checking out content without subscribing to a feed or saving
anything to your local cache.`,
	Run: func(cmd *cobra.Command, args []string) {
		fmt.Println("preview called")
	},
}

func init() {
	rootCmd.AddCommand(previewCmd)

	// Here you will define your flags and configuration settings.

	// Cobra supports Persistent Flags which will work for this command
	// and all subcommands, e.g.:
	// previewCmd.PersistentFlags().String("foo", "", "A help for foo")

	// Cobra supports local flags which will only run when this command
	// is called directly, e.g.:
	// previewCmd.Flags().BoolP("toggle", "t", false, "Help message for toggle")
}
