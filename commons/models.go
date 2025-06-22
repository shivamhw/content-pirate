package commons

var (
	IMG_SUFFIX = []string{".jpg", ".jpeg", ".png", ".gif"}
	VID_SUFFIX = []string{".mp4"}
)

type MediaType = string

const (
	VID_TYPE MediaType = "vids"
	IMG_TYPE MediaType = "imgs"
)

type Job struct {
	Id        string
	Src       string
	Dst       string
	Title     string
	FileName  string
	MediaType string
	SrcAcId   string
}
