package telegram

import (
	"context"
	"fmt"
	log "log/slog"

	"github.com/gotd/td/session"
	"github.com/gotd/td/telegram/auth"
	"github.com/gotd/td/tg"
	"github.com/iyear/tdl/core/tclient"
)

type LoginOpts struct {
	Phone string
	Otp   string
	Hash  string
	SessionPath string
	Force bool
}

func Login(opts *LoginOpts) (hash string, err error) {
	if opts.Otp == "" {
		log.Info("running send code flow", "user", opts.Phone)
		hash, err = SendCode(opts)
		if err != nil {
			return
		}
	} else {
		log.Info("running submit code flow", "user", opts)
		err = SubmitCode(opts)
		if err != nil {
			return
		}
	}
	return
}

// todo add support for other responses for sendCode
func SendCode(opts *LoginOpts) (hash string, err error) {
	client, err := tclient.New(context.Background(), tclient.Options{
		AppID:            Appid,
		AppHash:          AppHash,
		Session: &session.FileStorage{Path: opts.SessionPath},
	})
	if err != nil {
		return 
	}
	err = client.Run(context.Background(), func(ctx context.Context) (err error) {
		a := client.Auth()
		ok, err := a.Status(ctx)
		if err != nil {
			return
		}
		if ok.Authorized && !opts.Force {
			log.Warn("already logged in")
			return
		}
		s, err := a.SendCode(ctx, opts.Phone, auth.SendCodeOptions{})
		if err != nil {
			log.Error("send code", "err", err)
			return
		}
		switch s := s.(type) {
		case *tg.AuthSentCode:
			hash = s.PhoneCodeHash
			log.Info("using hash", "hash", hash)
			return
		}
		return
	})
	return
}

// todo add support for password
func SubmitCode(opts *LoginOpts) error {
	client, err := tclient.New(context.Background(), tclient.Options{
		AppID:            Appid,
		AppHash:          AppHash,
		Session:          &session.FileStorage{Path: opts.SessionPath},
	})
	if err != nil {
		return err
	}
	
	return client.Run(context.Background(), func(ctx context.Context) error {
		a := client.Auth()
		ok, err := a.Status(ctx)
		if err != nil {
			return err
		}
		if ok.Authorized {
			fmt.Print("already loggin h")
			return nil
		}
		log.Info("signin", "otp", opts.Otp, "hash", opts.Hash)
		_, err = a.SignIn(ctx, opts.Phone, opts.Otp, opts.Hash)
		if err != nil {
			log.Error("otp submit", "err", err)
			return err
		}
		log.Info("login success", "user", opts.Phone)
		return nil
	})
}
