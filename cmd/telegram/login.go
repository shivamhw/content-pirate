package telegram_cmd

import (
	"fmt"

	"github.com/shivamhw/content-pirate/pkg/telegram"
	"github.com/spf13/cobra"
)

//todo how to preapply telegram logins
func loginCmd() *cobra.Command {
	var loginOpts telegram.LoginOpts
	cmd := &cobra.Command{
		Use: "login",
		RunE: func(cmd *cobra.Command, args []string) error {
			if user.Phone == "" && loginOpts.Otp == ""{
				return fmt.Errorf("please enter either phone nm or otp")
			}
			fmt.Println("starting login flow")
			hash, err := telegram.Login(&telegram.LoginOpts{
				Phone: user.Phone,
				Otp: loginOpts.Otp,
				Hash: loginOpts.Hash,
				SessionPath: loginOpts.SessionPath,
				Force: true,
			})
			if hash != "" {
				fmt.Printf("hash %s", hash)
			}
			if err != nil {
				return err
			}
			return nil
		},
	}
	cmd.Flags().StringVar(&user.Phone, "phone", "", "phone nm of telegram")
	cmd.Flags().StringVar(&loginOpts.Otp, "otp", "", "otp for login")
	cmd.Flags().StringVar(&loginOpts.Hash, "hash", "", "hash for login")
	cmd.Flags().StringVar(&loginOpts.SessionPath, "session-path", "", "path to store session")
	return cmd
}
