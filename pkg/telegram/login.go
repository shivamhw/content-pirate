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
	if opts.Otp == "" {
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
	return t.c.Run(t.ctx, func(ctx context.Context) error {
		a := t.c.Auth()
		ok, err := a.Status(ctx)
		if err != nil {
			return err
		}
		if ok.Authorized {
			log.Warn("already logged in")
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
			log.Info("using hash", "hash", hash)
			return nil
		}
		return nil
	})
}

//todo add support for password
func (t *Telegram) Otp(opts *LoginOpts) error {
	return t.c.Run(t.ctx, func(ctx context.Context) error {
		a := t.c.Auth()
		ok, err := a.Status(ctx)
		if err != nil {
			return err
		}
		if ok.Authorized {
			fmt.Print("already loggin h")
			return nil
		}
		hash, err := t.user.Store.Kvd.Get(ctx, "codeHash")
		if err != nil {
			return fmt.Errorf("hash not found for %s", opts.Phone)
		}
		opts.Hash = string(hash)

		log.Info("signin","otp", opts.Otp,"hash", opts.Hash)
		_, err = a.SignIn(ctx, opts.Phone, opts.Otp, opts.Hash)
		if err != nil {
			log.Error("otp submit", "err", err)
			return err
		}
		return nil
	})
}