package store

import "github.com/shivamhw/content-pirate/commons"

type DstPath struct {
	BasePath     string
	CleanOnStart bool
	Type 		 string
}

type Store interface {
	Write(i *commons.Item) (string, error)
	ItemExists(i *commons.Item) bool
	DirExists(string) bool
	CreateDir(string) error
	CleanAll(string) error
}


func GetStore(d *DstPath) (Store, error) {
	if store, err := NewFileStore(d); err != nil {
		return nil, err
	} else {
		return store, nil
	}
}