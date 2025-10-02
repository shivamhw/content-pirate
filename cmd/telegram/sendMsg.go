package telegram_cmd

import (
	"context"
	"fmt"

	"github.com/shivamhw/content-pirate/pkg/telegram"
	"github.com/spf13/cobra"
)


//todo how to preapply telegram logins
func sendMsgCmd() *cobra.Command {
	var to string
	cmd := &cobra.Command{
		Use: "send",
		RunE: func(cmd *cobra.Command, args []string) error {
			t, err := telegram.NewTelegram(context.Background(), &user)
			if err != nil {
				return err
			}
			if st, _ := t.WhoAmI(); !st.Authorized {
				return fmt.Errorf("user is not authorized %s", user.Phone)
			}

			chats, err := t.SendMsg(to, "test msg")
			if err != nil {
				return err
			}
			fmt.Print(chats)
			return nil
		},
	}
	cmd.Flags().StringVar(&to, "to", "", "chat id")
	cmd.Flags().StringVar(&user.Phone, "phone", "", "phone nm of telegram")
	return cmd
}
