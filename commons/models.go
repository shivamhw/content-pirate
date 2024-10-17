package commons

import "path/filepath"



var (
	IMG_SUFFIX = []string{".jpg", ".jpeg", ".png", ".gif"}
	VID_SUFFIX = []string{".mp4"}
)

const (
	IMGS = "imgs"
	VIDS = "vids"
)


type Job struct {
	Src  string
	Dst  string
	Name string
}

type Post struct {
	MediaType string
	SrcLink   string
	Title     string
	Id        string
	SourceAc  string
	Ext       string
}

type DstPath struct {
	ImgPath    string
	VidPath    string
	BasePath   string
	CombineDir bool
}

func (d *DstPath) GetBasePath() string {
	return d.BasePath
}

func (d *DstPath) GetImgPath(r string) string {
	if d.CombineDir {
		return filepath.Join(d.BasePath, d.ImgPath)
	}
	return filepath.Join(d.BasePath, r, d.ImgPath)
}

func (d *DstPath) GetVidPath(r string) string {
	if d.CombineDir {
		return filepath.Join(d.BasePath, d.VidPath)
	}
	return filepath.Join(d.BasePath, r, d.VidPath)
}
