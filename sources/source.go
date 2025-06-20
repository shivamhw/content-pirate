package sources

import (
	"time"

	"github.com/shivamhw/reddit-pirate/commons"
)

type ScrapeOpts struct {
	Limit int
	Page  int
	Last  string
	Duration string
	LastFrom time.Time
	NextPage string
}

type Source interface {
	ScrapePosts(string, ScrapeOpts) (chan commons.Post, error)
	DownloadJob(commons.Job) ([]byte, error)
}
