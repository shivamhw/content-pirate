package telegram

import (
	"context"
	"os"
	"path/filepath"

	flow "github.com/gotd/td/telegram/auth"
)

func writePhoneNumToFile(phoneNumber string, basePath string) error {
	filePath := filepath.Join(basePath, "phone")
	return os.WriteFile(filePath, []byte(phoneNumber), 0644)
}

func (t *Telegram) LoginWithCode(ctx context.Context, user *UserData) error {
	store, err:= t.GetStorage(user)
	if err != nil {	
		return err
	}
	c, err := t.GetClientWithStore(user, store, false)
	if err != nil {
		return err
	}
	writePhoneNumToFile(user.PhoneNumber, t.users[user.PhoneNumber].Session.BasePath)

	return c.Run(ctx, func(ctx context.Context) error {
		if err = c.Ping(ctx); err != nil {
			return err
		}

		flow := flow.NewFlow(File(store.BasePath), flow.SendCodeOptions{})
		if err = c.Auth().IfNecessary(ctx, flow); err != nil {
			return err
		}

		_, err := c.Self(ctx)
		if err != nil {
			return err
		}

		return nil
	})
}
