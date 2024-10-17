package store

import "io"

type Store interface {
	Write(path string, data io.Reader) error
	DirExists(string) bool
	CreateDir(string)
	CleanAll(string) error
}