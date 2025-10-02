package telegram_cmd

import (
	"context"
	"fmt"

	"github.com/shivamhw/content-pirate/pkg/telegram"
	"github.com/spf13/cobra"
)

var user telegram.ClientOpts

func lsCmd() *cobra.Command {
	cmd := &cobra.Command{
		Use: "chats",
		RunE: func(cmd *cobra.Command, args []string) error {
			t, err := telegram.NewTelegram(context.Background(), &telegram.ClientOpts{
				Phone: user.Phone,
			})
			if err != nil {
				return err
			}
			if st, _ := t.WhoAmI(); !st.Authorized {
				return fmt.Errorf("user is not authorized %s", user.Phone)
			}

			chats, err := t.ListChats()
			if err != nil {
				return err
			}
			for _, c := range chats {
				fmt.Printf("%s | %d | %d \n", c.VisibleName, c.ID, c.AccessHash)
			}
			return nil
		},
	}
	cmd.Flags().StringVar(&user.Phone, "phone", "", "phone nm of telegram")
	return cmd
}
