package main

import (
	"context"

	"github.com/shivamhw/reddit-pirate/pkg/telegram"
)

func main() {
	// cmd.Execute()
	// tdl.Code(context.Background())
	ctx := context.Background()
	t := telegram.NewTelegram(ctx)
	user :=  &telegram.UserData{PhoneNumber: "+918085026377"}
	// t.ListChats(&telegram.UserData{PhoneNumber: "+918085026377"})
	// t.ExportChat(&telegram.UserData{PhoneNumber: "+918085026377",}, telegram.ExportOpts{
	// 	ChatId: "1237061921",
	// 	Limit: 100,
	// })
	t.DownloadExport(user, telegram.DownloadOpts{})
}
