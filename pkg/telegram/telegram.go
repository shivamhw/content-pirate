package telegram

import (
	"context"
	"fmt"
	"log/slog"

	"github.com/gotd/td/telegram"
	"github.com/iyear/tdl/app/chat"
	"github.com/iyear/tdl/core/logctx"
	tclientcore "github.com/iyear/tdl/core/tclient"
	"github.com/iyear/tdl/pkg/tclient"
)


type Telegram struct {
	users map[string]*UserData
	ctx context.Context
}

type UserData struct {
	PhoneNumber string
	Session     *Store
}


func NewTelegram(ctx context.Context) *Telegram {
	return &Telegram{
		users: make(map[string]*UserData),
		ctx: ctx,
	}
}

func (t *Telegram) Login(opts *UserData, clean bool) error {
	store, err := NewStore(t.ctx, opts.PhoneNumber, clean)
	defer func() {	
		if err != nil {
			store.Close()
		}
	}()

	if err != nil {
		return err
	}
	t.users[opts.PhoneNumber] = &UserData{
		PhoneNumber: opts.PhoneNumber,
		Session: store,
	}
	t.LoginWithCode(t.ctx, opts)
	return nil
}
 
func (t *Telegram) GetStorage(opts *UserData) (*Store, error) {
	user, ok := t.users[opts.PhoneNumber]
	if !ok {
		slog.Error("user not found, creating one", "phone", opts.PhoneNumber)
		store, err := GetOrCreateStore(t.ctx, opts.PhoneNumber)
		if err != nil {
			return nil, err
		}
		return store, nil
	}
	return user.Session, nil
}

func (t *Telegram) GetClientWithStore(opts *UserData, store *Store, login bool) (*telegram.Client, error) {
	c, err := tclient.New(t.ctx, tclient.Options{
		KV:               store.Kvd,
		UpdateHandler:    nil,
	}, login)
	if err != nil {
		return nil, err
	}
	return c, nil
}

func (t *Telegram) ListChats(opts *UserData) error {
	var result []*chat.Dialog
	store, err:= t.GetStorage(opts)
	if err != nil {	
		return err
	}
	c, err := t.GetClientWithStore(opts, store, false)
	if err != nil {
		return err
	}
	err = tclientcore.RunWithAuth(t.ctx, c, func(ctx context.Context) error {
		result, err = chat.List(logctx.Named(ctx, "ls"), c, store.Kvd, chat.ListOptions{Filter: "true"})
		return err
	})
	if err != nil {
		return err
	}
	for _, r := range result {
		fmt.Println(r.VisibleName)
	}
	return nil
}
