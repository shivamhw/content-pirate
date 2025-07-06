package telegram_cmd

import (
	"context"
	"fmt"

	"github.com/shivamhw/content-pirate/pkg/telegram"
	"github.com/spf13/cobra"
)


var (
	to telegram.Recipient
)
//todo how to preapply telegram logins
func sendMsgCmd() *cobra.Command {
	cmd := &cobra.Command{
		Use: "send",
		RunE: func(cmd *cobra.Command, args []string) error {
			t, err := telegram.NewTelegram(context.Background(), &user)
			if err != nil {
				return err
			}
			if st, _ := t.WhoAmI(); !st.Authorized {
				return fmt.Errorf("user is not authorized %s", user.PhoneNumber)
			}

			chats, err := t.SendMsg(&to, "test msg")
			if err != nil {
				return err
			}
			fmt.Print(chats)
			return nil
		},
	}
	cmd.Flags().Int64Var(&to.UserId, "user", int64(0), "chat id")
	cmd.Flags().StringVar(&user.PhoneNumber, "phone", "", "phone nm of telegram")
	return cmd
}
