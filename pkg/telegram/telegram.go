package telegram

import (
	"context"
	"fmt"
	"log/slog"
	"path/filepath"
	"time"

	"github.com/gotd/td/telegram"
	"github.com/iyear/tdl/app/chat"
	"github.com/iyear/tdl/app/dl"
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
		ReconnectTimeout: 5*time.Minute,
	}, login)
	if err != nil {
		return nil, err
	}
	return c, nil
}

func (t *Telegram) ListChats(opts *UserData) error {
	var result []*chat.Dialog 
	resltCtx := context.WithValue(t.ctx, "results", &result)
	store, err:= t.GetStorage(opts)
	if err != nil {	
		return err
	}
	c, err := t.GetClientWithStore(opts, store, false)
	if err != nil {
		return err
	}
	err = tclientcore.RunWithAuth(resltCtx, c, func(ctx context.Context) error {
		err = chat.List(logctx.Named(ctx, "ls"), c, store.Kvd, chat.ListOptions{Filter: "true"})
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

type ExportOpts struct {
	ChatId string
	Limit int
}

func (t *Telegram) ExportChat(user *UserData, opts ExportOpts) error {
	exprtOpts := chat.ExportOptions{
		Chat: opts.ChatId,
		Type: chat.ExportTypeLast,
		Input: []int{opts.Limit},
		Filter: "true",
		OnlyMedia: true,
	}
	store, err:= t.GetStorage(user)
	if err != nil {	
		return err
	}
	exprtOpts.Output = filepath.Join(store.BasePath, opts.ChatId)
	c, err := t.GetClientWithStore(user, store, false)
	if err != nil {
		return err
	}
	err = tclientcore.RunWithAuth(t.ctx, c, func(ctx context.Context) error {
		return chat.Export(ctx, c, store.Kvd, exprtOpts)
	})
	return err
}

type DownloadOpts struct {
	ChatId string
}

func (t *Telegram) DownloadExport(user *UserData, opts DownloadOpts) error {
	dwnldOpts := dl.Options{
		Files: []string{"/home/shivamhw/Code/reddit-pirate/Rdata/+918085026377/1237061921"},
		Continue: true,
		Template: "{{ .DialogID }}_{{ .MessageID }}_{{ filenamify .FileName }}",
	}
	store, err:= t.GetStorage(user)
	if err != nil {	
		return err
	}
	dwnldOpts.Dir = filepath.Join(store.BasePath, opts.ChatId+"test", "files")
	c, err := t.GetClientWithStore(user, store, false)
	if err != nil {
		return err
	}
	err = tclientcore.RunWithAuth(t.ctx, c, func(ctx context.Context) error {
		return dl.Run(ctx, c, store.Kvd, dwnldOpts)
	})
	return err
}
