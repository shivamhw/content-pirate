package telegram_cmd


import "github.com/spf13/cobra"

var SessionPath string 

func TelegramCmd() *cobra.Command {
	var cmd = cobra.Command{
		Use : "telegram", 
		Short: "telegram specific cmds",
	}
	cmd.PersistentFlags().StringVar(&SessionPath, "session-path", "./session.json", "set telegram session")
	cmd.MarkFlagRequired("session-path")
	cmd.AddCommand(lsCmd())
	cmd.AddCommand(sendMsgCmd())
	cmd.AddCommand(scrapeCmd())
	cmd.AddCommand(loginCmd())
	
	return &cmd
}

