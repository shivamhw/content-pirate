package reddit_cmd

import (
	"context"
	"fmt"

	"github.com/shivamhw/content-pirate/pkg/reddit"
	"github.com/spf13/cobra"
)

type UserCfg struct {
	authCfg string
}

var uCfg UserCfg

func userCmd() *cobra.Command {
	var cmd = &cobra.Command{
		Use:   "user",
		Short: "A brief description of your application",
		RunE:  userCmdInit,
	}
	cmd.Flags().StringVar(&uCfg.authCfg, "auth", "./reddit.json", "auth config for reddit")
	return cmd
}

func userCmdInit(cmd *cobra.Command, args []string) error {
	r, err := reddit.NewRedditClient(context.Background(), reddit.RedditClientOpts{
		CfgPath: uCfg.authCfg,
	})
	if err != nil {
		return err
	}
	res, err := r.GetSubscribedSubreddits(100)
	if err != nil {
		return err
	}
	for _, sub := range res {
		fmt.Printf("subname %s\n", sub.Name)
	}
	return nil
}
