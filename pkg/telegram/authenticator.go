package telegram

import (
	"context"
	"fmt"
	"os"
	"path/filepath"
	flow "github.com/gotd/td/telegram/auth"
	"github.com/gotd/td/tg"
	"time"
)

var codeChan = make(chan struct{
	code string
	phone string
})

type fileAuth struct {
	basePath string
	flow.CodeAuthenticator
}

func File(basePath string) flow.UserAuthenticator {
	return fileAuth{
		basePath:          basePath,
		CodeAuthenticator: getFileCode(basePath),
	}
}

func (f fileAuth) Phone(ctx context.Context) (string, error) {
	phone, err := os.ReadFile(filepath.Join(f.basePath, "phone"))
	if err != nil {
		return "", fmt.Errorf(err.Error())
	}
	return string(phone), nil
}

func (f fileAuth) Password(ctx context.Context) (string, error) {
	password, err := os.ReadFile(filepath.Join(f.basePath, "password"))
	if err != nil {
		return "", fmt.Errorf(err.Error())
	}
	return string(password), nil
}

func getFileCode(basePath string) flow.CodeAuthenticator {
	return flow.CodeAuthenticatorFunc(func(ctx context.Context, sentCode *tg.AuthSentCode) (string, error) {
		for {
			_, err := os.Stat(filepath.Join(basePath, "code"))
			if err != nil {
				time.Sleep(1 * time.Second)
				fmt.Println("Waiting for code file to be created...")
				continue
			} else {
				break
			}
		}
		code, err := os.ReadFile(filepath.Join(basePath, "code"))
		if err != nil {
			return "", fmt.Errorf(err.Error())
		}
		return string(code), nil
	})
}

func (f fileAuth) AcceptTermsOfService(ctx context.Context, tos tg.HelpTermsOfService) error {
	return nil
}


func (f fileAuth) SignUp(ctx context.Context) (flow.UserInfo, error) {
	return flow.UserInfo{
		FirstName: "Test",
		LastName:  "User",
	}, nil
}