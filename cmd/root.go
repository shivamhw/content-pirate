package cmd

import (
	"os"

	reddit_cmd "github.com/shivamhw/content-pirate/cmd/reddit"
	"github.com/spf13/cobra"
)

// rootCmd represents the base command when called without any subcommands
var rootCmd = &cobra.Command{
	Use:   "content-pirate",
	Short: "A brief description of your application",
}

func Execute() {
	err := rootCmd.Execute()
	if err != nil {
		os.Exit(1)
	}
}

func init() {

	rootCmd.AddCommand(helloCmd)
	rootCmd.AddCommand(reddit_cmd.RedditCmd())

	rootCmd.Flags().BoolP("toggle", "t", false, "Help message for toggle")
}
