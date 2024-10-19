package cmd

import (
	"context"
	"fmt"
	"log"

	. "github.com/shivamhw/reddit-pirate/commons"
	"github.com/spf13/cobra"
	"github.com/vartanbeno/go-reddit/v2/reddit"
)

type userClient struct {
	reddit *reddit.Client
}

var userCmd = &cobra.Command{
	Use:   "user",
	Short: "A brief description of your application",
	Long: `A longer description that spans multiple lines and likely contains
examples and usage of using your application. For example:

Cobra is a CLI library for Go that empowers applications.
This application is a tool to generate the needed files
to quickly create a Cobra application.`,
	// Uncomment the following line if your bare application
	// has an action associated with it:
	// Run: func(cmd *cobra.Command, args []string) { },
	Run: userCmdInit,
}

type UserCfg struct {
	authCfg string
}

var uCfg UserCfg

func init() {
	userCmd.Flags().StringVar(&uCfg.authCfg, "auth", "./reddit.json", "auth config for reddit")
}

func userCmdInit(cmd *cobra.Command, args []string) {
	ReadFromJson(uCfg.authCfg, &aCfg)
	credentials := reddit.Credentials{ID: aCfg.ID, Secret: aCfg.Secret, Username: aCfg.Username, Password: aCfg.Password}
	c, err := reddit.NewClient(credentials)
	if err != nil {
		log.Fatalf("err creating client %s", err)
	}
	client := userClient{
		reddit: c,
	}

	nextToken := ""
	for {
		subs, resp, err := client.reddit.Subreddit.Subscribed(context.Background(), &reddit.ListSubredditOptions{
			ListOptions: reddit.ListOptions{
				Limit: 100,
				After: nextToken,
			},
		})
		if err != nil {
			log.Fatalf("err %s", err)
		}
		nextToken = resp.After
		for _, s := range subs {
			fmt.Println(s.Name)
		}
		if nextToken == "" {
			break
		}
	}
}
