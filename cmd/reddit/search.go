package reddit_cmd

import (
	"context"
	"fmt"
	"github.com/shivamhw/reddit-pirate/pkg/reddit"
	"github.com/spf13/cobra"
)

func searchCmd() *cobra.Command {
	var query string
	var limit int
	var cmd = &cobra.Command{
		Use:   "search",
		Short: "A brief description of your application",
		RunE: func(cmd *cobra.Command, args []string) error {
			r, err := reddit.NewRedditClient(context.Background(), reddit.RedditClientOpts{
				CfgPath: uCfg.authCfg,
			})
			if err != nil {
				return err
			}
			res, err := r.SearchSubreddits(query, limit)
			if err != nil {
				return err
			}
			for _, sub := range res {
				fmt.Printf("subname %s\n", sub.Name)
			}
			return nil
		},
	}
	cmd.Flags().StringVar(&uCfg.authCfg, "auth", "./reddit.json", "auth config for reddit")
	cmd.Flags().StringVar(&query, "query", "", "search term")
	cmd.Flags().IntVar(&limit, "limit", 10, "limit")
	cmd.MarkFlagRequired("query")
	return cmd
}
