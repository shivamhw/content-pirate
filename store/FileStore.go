package store

import (
	"fmt"
	"os"
	"path/filepath"

	"github.com/shivamhw/content-pirate/commons"
	. "github.com/shivamhw/content-pirate/pkg/log"
)

type FileDstPath struct {
	BasePath string
	Clean    bool
}

func (f FileDstPath) GetBasePath() string {
	return f.BasePath
}

func (f FileDstPath) CleanOnStart() bool {
	return f.Clean
}

func (f FileDstPath) Type() DstPathType {
	return FILE_DST_PATH
}

type FileStore struct {
	Dst *FileDstPath
}

func NewFileStore(path *FileDstPath) (*FileStore, error) {
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
	fullFilePath := fmt.Sprintf("%s/%s", i.Dst, i.FileName)
	f.CreateDir(filepath.Dir(fullFilePath))
	outfile, err := os.Create(fullFilePath)
	if err != nil {
		return "", err
	}
	defer outfile.Close()
	_, err = outfile.Write(i.Data)
	return fullFilePath, err
}

func (f *FileStore) CreateDir(path string) (err error) {
	return os.MkdirAll(path, 0755)
}

func (f *FileStore) ID() (string) {
	return f.Dst.BasePath
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

func (d *FileDstPath) sanitize() (err error) {
	if d.BasePath == "" {
		d.BasePath = "./download"
	}
	d.BasePath, err = filepath.Abs(d.BasePath)
	Logger.Info("download path", "path", d.BasePath)
	return err
}

func (f *FileStore) createStructure() (err error) {
	if f.Dst.Clean {
		err := f.CleanAll(f.Dst.BasePath)
		if err != nil {
			Logger.Warn("err while deleting dir structure ", "error", err)
		} else {
			Logger.Info("cleanup success")
		}
	}
	return f.CreateDir(f.Dst.BasePath)
}

func (f *FileStore) ItemExists(i *commons.Item) bool {
	path := fmt.Sprintf("%s/%s", f.Dst.BasePath, i.FileName)
	if _, err := os.Stat(path); err != nil {
		return false
	}
	return true
}

func (f *FileStore) GetItemDstPath(i *commons.Item) string {
	return fmt.Sprintf("%s/%s", f.Dst.BasePath, i.SourceAc)
}
