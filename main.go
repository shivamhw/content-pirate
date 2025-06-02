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
	t.ListChats(&telegram.UserData{PhoneNumber: "+918085026377"})
}
