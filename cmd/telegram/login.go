package telegram_cmd

import (
	"context"

	"github.com/shivamhw/content-pirate/pkg/telegram"
	"github.com/spf13/cobra"
)

//todo how to preapply telegram logins
func loginCmd() *cobra.Command {
	var otp string
	cmd := &cobra.Command{
		Use: "login",
		RunE: func(cmd *cobra.Command, args []string) error {
			t, err := telegram.NewTelegram(context.Background(), &user)
			if err != nil {
				return err
			}

			err = t.Login(&telegram.LoginOpts{
				Phone: user.PhoneNumber,
				Otp: otp,
			}, false)
			if err != nil {
				return err
			}
			return nil
		},
	}
	cmd.Flags().StringVar(&user.PhoneNumber, "phone", "", "phone nm of telegram")
	cmd.Flags().StringVar(&otp, "otp", "", "otp for login")
	return cmd
}
