package store

import (
	"log"
	"os"
)

type FileStore struct {
	Dir string
}

func (f FileStore) DirExists(path string) bool {
	info, err := os.Stat(path)
	if os.IsNotExist(err) {
		return false
	}
	return info.IsDir()
}

func (f FileStore) Write(path string, data []byte) (err error) {
	outfile, err := os.Create(path)
	if err != nil {
		return err
	}
	defer outfile.Close()
	_, err = outfile.Write(data)
	return err
}

func (f FileStore) CreateDir(path string) {
	os.MkdirAll(path, 0755)
}

func (f FileStore) CleanAll(path string) error {
	err := os.RemoveAll(path)
	if err != nil {
		log.Print("err while deleting dir structure ", err)
	} else {
		log.Print("cleanup success")
	}
	return err
}
