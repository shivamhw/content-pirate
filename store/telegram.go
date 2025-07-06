package store

import (
	"fmt"

	"github.com/shivamhw/content-pirate/commons"
	. "github.com/shivamhw/content-pirate/pkg/log"
	"github.com/shivamhw/content-pirate/pkg/telegram"
)

type TelegramStore struct {
	cfg *TelegramDstPath
	C   *telegram.Telegram
}

type TelegramDstPath struct {
	ChatId   int
	PhoneNumber    string
}

func NewTelegramStore(cfg *TelegramDstPath) (*TelegramStore, error) {
	return &TelegramStore{
		cfg: cfg,
	}, nil
}

func (s *TelegramStore) CreateDir(d string) error {
	return nil
}

func (s *TelegramStore) CleanAll(d string) error {
	return nil
}

func (s *TelegramStore) ItemExists(i *commons.Item) bool {
	return false
}

func (s *TelegramStore) ID() string {
	return fmt.Sprintf("%d", s.cfg.ChatId)
}

func (s *TelegramStore) GetItemDstPath(i *commons.Item) string {
	return fmt.Sprintf("%d", s.cfg.ChatId)
}

func (s *TelegramStore) Write(i *commons.Item) (path string, err error) {
	_, err = s.C.ForwardMsg(i.SourceAc, i.Dst, i.Id)
	if err != nil {
		return "", err
	}
	Logger.Info("forwarded msg", "from", i.SourceAc, "to", i.Dst, "msg", i.FileName)
	return i.Dst, err
}

func (t TelegramDstPath) GetBasePath() string {
	return fmt.Sprintf("%d", t.ChatId)
}

func (t TelegramDstPath) CleanOnStart() bool {
	return false
}

func (t TelegramDstPath) Type() DstPathType {
	return TELEGRAM_DST_PATH
}
