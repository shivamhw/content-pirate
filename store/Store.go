package store

import "github.com/shivamhw/content-pirate/commons"


type Store interface {
	Write(path string,t commons.MediaType, data []byte) (string, error)
	DirExists(string) bool
	CreateDir(string)
	CleanAll(string) error
}