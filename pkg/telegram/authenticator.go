package telegram

import (
	"context"
	"log/slog"
	"time"

	flow "github.com/gotd/td/telegram/auth"
	"github.com/gotd/td/tg"
)
type asyncAuth struct {
	phone string
	code string
	password string
}

func NewAsyncAuth() flow.UserAuthenticator {
	return &asyncAuth{}
}

func (f *asyncAuth) Code(ctx context.Context, sentCode *tg.AuthSentCode) (string, error){
	for {
		if f.code != "" {
			break
		} else {
			slog.Info("waiting for code")
		}
		time.Sleep(3*time.Second)
	}
	return string(f.code), nil
}

func (f *asyncAuth) Phone(ctx context.Context) (string, error) {
	for {
		if f.phone != "" {
			slog.Info("auth", "phone number found ", f.phone)
			break
		} else {
			slog.Info("waiting for phone num")
		}
		time.Sleep(3*time.Second)
	}
	return string(f.phone), nil
}

func (f *asyncAuth) Password(ctx context.Context) (string, error) {
	for {
		if f.password != "" {
			break
		} else {
			slog.Info("waiting for password")
		}
		time.Sleep(10*time.Second)
	}
	return string(f.password), nil
}


func (f *asyncAuth) AcceptTermsOfService(ctx context.Context, tos tg.HelpTermsOfService) error {
	return nil
}


func (f *asyncAuth) SignUp(ctx context.Context) (flow.UserInfo, error) {
	return flow.UserInfo{
		FirstName: "Test",
		LastName:  "User",
	}, nil
}

func (f *asyncAuth) SetPhone(phone string) {
	f.phone = phone
}

func (f *asyncAuth) SetCode(code string) {
	f.code = code
}

func (f *asyncAuth) SetPass(pass string) {
	f.password = pass
}