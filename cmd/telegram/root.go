package telegram_cmd


import "github.com/spf13/cobra"

func TelegramCmd() *cobra.Command {
	var cmd = cobra.Command{
		Use : "telegram", 
		Short: "telegram specific cmds",
	}
	cmd.AddCommand(lsCmd())
	cmd.AddCommand(sendMsgCmd())
	cmd.AddCommand(scrapeCmd())
	cmd.AddCommand(loginCmd())
	return &cmd
}

