package reddit_cmd

import "github.com/spf13/cobra"

func RedditCmd() *cobra.Command {
	var cmd = cobra.Command{
		Use : "reddit", 
		Short: "reddit specific cmds",
	}
	cmd.AddCommand(userCmd())
	cmd.AddCommand(searchCmd())
	return &cmd
}
