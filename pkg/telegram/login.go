package telegram

import (
	"context"
	"fmt"
	log "log/slog"

	"github.com/gotd/td/telegram/auth"
	"github.com/gotd/td/tg"
)

type LoginOpts struct {
	Phone string
	Otp   string
	Hash  string
}



func (t *Telegram) Login(opts *LoginOpts, force bool) error {
	var store *Store
	var err error
	if force {
		store, err = NewStore(t.ctx, opts.Phone, force)
		if err != nil {
			log.Error("store", err)
			return err
		}
	} else {
		store, err = GetOrCreateStore(t.ctx, opts.Phone)
		if err != nil {
			return err
		}
	}

	t.user = &UserData{
		PhoneNumber: opts.Phone,
		Store:       store,
	}
	if opts.Hash == "" {
		log.Info("running send code flow", "user", opts.Phone)
		err := t.SendCode(opts)
		if err != nil {
			return err
		}
	} else {
		log.Info("running submit code flow", "user", opts)
		err := t.Otp(opts)
		if err != nil {
			return err
		}
	}
	return nil
}

//todo add support for other responses for sendCode
func (t *Telegram) SendCode(opts *LoginOpts) error {
	t.c.Run(t.ctx, func(ctx context.Context) error {
		a := t.c.Auth()
		ok, err := a.Status(ctx)
		if err != nil {
			return err
		}
		if ok.Authorized {
			fmt.Print("already logged in")
			return nil
		}
		s, err := a.SendCode(ctx, opts.Phone, auth.SendCodeOptions{})
		if err != nil {
			log.Error("send code" ,"err", err)
			return err
		}
		switch s := s.(type) {
		case *tg.AuthSentCode:
			hash := s.PhoneCodeHash
			log.Info(hash)
			t.user.Store.Kvd.Set(ctx, "codeHash", []byte(hash))
			fmt.Printf("using hash %s", hash)
			return nil
		}

		return nil
	})
	return fmt.Errorf("not valid opts for login")
}

//todo add support for password
func (t *Telegram) Otp(opts *LoginOpts) error {
	t.c.Run(t.ctx, func(ctx context.Context) error {
		a := t.c.Auth()
		ok, err := a.Status(ctx)
		if ok.Authorized {
			fmt.Print("already loggin h")
			return nil
		}
		fmt.Printf("using code %s for hash %s", opts.Otp, opts.Hash)
		_, err = a.SignIn(ctx, opts.Phone, opts.Otp, opts.Hash)
		if err != nil {
			log.Error("otp submit", "err", err)
			return err
		}
		return nil
	})
	return fmt.Errorf("not valid opts for login")
}