package telegram

import (
	"context"
	"sync"

	flow "github.com/gotd/td/telegram/auth"
)

func (t *Telegram) LoginWithCode(ctx context.Context, user *UserData) error {
	store, err:= t.GetStorage(user)
	if err != nil {	
		return err
	}
	c, err := t.GetClientWithStore(user, store, true)
	if err != nil {
		return err
	}
	auth := NewAsyncAuth()
	var wg sync.WaitGroup
	wg.Add(1)
	go func(wg *sync.WaitGroup) {
		defer wg.Done()
		c.Run(ctx, func(ctx context.Context) error {
		if err = c.Ping(ctx); err != nil {
			return err
		}

		flow := flow.NewFlow(auth, flow.SendCodeOptions{})
		if err = c.Auth().IfNecessary(ctx, flow); err != nil {
			return err
		}
		_, err := c.Self(ctx)
		if err != nil {
			return err
		}

		return nil
	})}(&wg)
	wg.Wait()
	return nil
}
