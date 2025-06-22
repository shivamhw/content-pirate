package store

import (
	"os"
	"path/filepath"
	"strings"

	"github.com/shivamhw/content-pirate/commons"
	. "github.com/shivamhw/content-pirate/pkg/log"
)

var DefaultPaths = map[string]string{
	commons.IMG_TYPE: "imgs",
	commons.VID_TYPE: "vids",
}

type DstPath struct {
	ImgPath      string
	VidPath      string
	BasePath     string
	CombineDir   bool
	CleanOnStart bool
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

func (f *FileStore) Write(path string, t commons.MediaType, data []byte) (string, error) {
	switch t {
	case commons.IMG_TYPE:
		path = filepath.Join(f.Dst.BasePath, f.Dst.ImgPath, path)
	case commons.VID_TYPE:
		path = filepath.Join(f.Dst.BasePath, f.Dst.VidPath, path)
	}
	if !f.Dst.CombineDir {
		f.CreateDir(filepath.Dir(path))
	}
	outfile, err := os.Create(path)
	if err != nil {
		return path, err
	}
	defer outfile.Close()
	_, err = outfile.Write(data)
	return path,err
}

func (f *FileStore) CreateDir(path string) {
	os.MkdirAll(path, 0755)
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

func (d *DstPath) sanitize() error {
	if d.BasePath == "" {
		d.BasePath = "./download"
	}
	if d.ImgPath == "" {
		d.ImgPath = filepath.Join(d.BasePath, DefaultPaths[commons.IMG_TYPE])
	}
	if d.VidPath == "" {
		d.VidPath = filepath.Join(d.BasePath, DefaultPaths[commons.VID_TYPE])
	}
	d.BasePath = strings.TrimSpace(d.BasePath)
	d.ImgPath = strings.TrimSpace(d.ImgPath)
	d.VidPath = strings.TrimSpace(d.VidPath)
	return nil
}

func (f *FileStore) createStructure() {
	if f.Dst.CleanOnStart {
		err := f.CleanAll(f.Dst.BasePath)
		if err != nil {
			Logger.Warn("err while deleting dir structure ", "error", err)
		} else {
			Logger.Info("cleanup success")
		}
	}
	Logger.Info("creating ", "path", f.Dst.ImgPath)
	Logger.Info("creating ", "path", f.Dst.VidPath)
	f.CreateDir(filepath.Join(f.Dst.BasePath, f.Dst.ImgPath))
	f.CreateDir(filepath.Join(f.Dst.BasePath, f.Dst.VidPath))
}
