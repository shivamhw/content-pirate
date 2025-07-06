package store

import (
	"fmt"

	"github.com/shivamhw/content-pirate/commons"
)

type DstPathType string

const (
	FILE_DST_PATH DstPathType = "FILE_DST_PATH"
	TELEGRAM_DST_PATH DstPathType = "TELEGRAM_DST_PATH"
)

type DstPath interface {
	GetBasePath() string
	CleanOnStart() bool
	Type() DstPathType
}

type Store interface {
	Write(i *commons.Item) (string, error)
	ItemExists(i *commons.Item) bool
	GetItemDstPath(i *commons.Item) string
	CreateDir(string) error
	CleanAll(string) error
	ID() string
}

func GetStore(d DstPath) (Store, error) {
	switch p := d.(type) {
	case TelegramDstPath:
		if store, err := NewTelegramStore(&p); err != nil {
			return nil, err
		} else {
			return store, nil
		}
	case FileDstPath:
		if store, err := NewFileStore(&p); err != nil {
			return nil, err
		} else {
			return store, nil
		}
	}
	return nil, fmt.Errorf("unknown dst store type")
}
