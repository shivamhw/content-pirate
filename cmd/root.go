package cmd

import (
	"log/slog"
	"os"

	reddit_cmd "github.com/shivamhw/content-pirate/cmd/reddit"
	telegram_cmd "github.com/shivamhw/content-pirate/cmd/telegram"
	"github.com/shivamhw/content-pirate/pkg/log"
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
	var lvl string

	rootCmd.AddCommand(helloCmd)
	rootCmd.AddCommand(reddit_cmd.RedditCmd())
	rootCmd.AddCommand(telegram_cmd.TelegramCmd())
	rootCmd.PersistentFlags().StringVarP(&lvl, "log-level", "l", "debug", "set log level")
	rootCmd.PersistentPreRun = func(cmd *cobra.Command, args []string) {
		switch lvl {
		case "debug":
			log.SetLevel(slog.LevelDebug)
		case "warn":
			log.SetLevel(slog.LevelWarn)
		case "info":
			log.SetLevel(slog.LevelInfo)
		case "err":
			log.SetLevel(slog.LevelError)
		default:
			log.SetLevel(slog.LevelDebug)
		}
	}
}
