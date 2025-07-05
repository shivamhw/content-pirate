package store

import (
	"fmt"

	"github.com/shivamhw/content-pirate/commons"
)

type DstPathType string

const (
	FILE_DST_PATH DstPathType = "FILE_DST_PATH"
)

type DstPath interface {
	GetBasePath() string
	CleanOnStart() bool
	Type() DstPathType
}

type Store interface {
	Write(i *commons.Item) (string, error)
	ItemExists(i *commons.Item) bool
	DirExists(string) bool
	CreateDir(string) error
	CleanAll(string) error
}

func GetStore(d DstPath) (Store, error) {
	switch p := d.(type) {
	case FileDstPath:
		if store, err := NewFileStore(&p); err != nil {
			return nil, err
		} else {
			return store, nil
		}
	}
	return nil, fmt.Errorf("unknown dst store type")
}
