package store

import (
	"fmt"
	"os"
	"path/filepath"

	"github.com/shivamhw/content-pirate/commons"
	. "github.com/shivamhw/content-pirate/pkg/log"
)

var DefaultPaths = map[string]string{
	commons.IMG_TYPE: "imgs",
	commons.VID_TYPE: "vids",
}


type FileStore struct {
	Dst *DstPath
}

func NewFileStore(path *DstPath) (*FileStore, error) {
	err := path.sanitize()
	if err != nil {
		return nil, err
	}
	f := &FileStore{
		Dst: path,
	}
	f.createStructure()
	return f, nil
}

func (f *FileStore) DirExists(path string) bool {
	info, err := os.Stat(path)
	if os.IsNotExist(err) {
		return false
	}
	return info.IsDir()
}

func (f *FileStore) Write(i *commons.Item) (string, error) {
	path := fmt.Sprintf("%s/%s", f.Dst.BasePath, i.Dst)
	f.CreateDir(filepath.Dir(path))
	outfile, err := os.Create(path)
	if err != nil {
		return path, err
	}
	defer outfile.Close()
	_, err = outfile.Write(i.Data)
	return path,err
}

func (f *FileStore) CreateDir(path string) (err error) {
	return os.MkdirAll(path, 0755)
}

func (f *FileStore) CleanAll(path string) error {
	err := os.RemoveAll(path)
	if err != nil {
		Logger.Error("err while deleting dir structure", "err", err)
	} else {
		Logger.Info("cleanup success")
	}
	return err
}

func (d *DstPath) sanitize() (err error) {
	if d.BasePath == "" {
		d.BasePath = "./download"
	}
	d.BasePath, err  = filepath.Abs(d.BasePath)
	Logger.Info("download path", "path", d.BasePath)
	return err
}

func (f *FileStore) createStructure() (err error) {
	if f.Dst.CleanOnStart {
		err := f.CleanAll(f.Dst.BasePath)
		if err != nil {
			Logger.Warn("err while deleting dir structure ", "error", err)
		} else {
			Logger.Info("cleanup success")
		}
	}
	return f.CreateDir(f.Dst.BasePath)
}

func (f *FileStore) ItemExists(i *commons.Item) (bool) {
	path := fmt.Sprintf("%s/%s", f.Dst.BasePath, i.Dst)
	if _, err := os.Stat(path); err != nil {
		return false
	}
	return true
}