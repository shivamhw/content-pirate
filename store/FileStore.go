package store

import (
	"io"
	"os"
)

type FileStore struct {
	Dir string
}

func (f FileStore) DirExists(path string) bool {
	info, err := os.Stat(path)
	if os.IsNotExist(err){
		return false
	}
	return info.IsDir()
}

func (f FileStore) Write(path string, src io.Reader) (err error) {
	outfile, err := os.Create(path)
	if err != nil {
		return err
	}
	defer outfile.Close()
	_, err = io.Copy(outfile, src)
	return err
}

func (f FileStore) CreateDir(path string){
	os.MkdirAll(path, 0755)
}