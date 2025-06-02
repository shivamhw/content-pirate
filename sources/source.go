package sources

import (
	"github.com/shivamhw/reddit-pirate/commons"
)

type Source interface {
	ScrapePosts(string, chan<- commons.Post)
	DownloadJob(commons.Job) ([]byte, error)
}
