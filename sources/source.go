package sources

import (
	"context"
	"time"

	"github.com/shivamhw/content-pirate/pkg/reddit"
)

type Post struct {
	MediaType string
	SrcLink   string
	Title     string
	Id        string
	SourceAc  string
	Ext       string
}

type ScrapeOpts struct {
	Limit          int
	Page           int
	Last           string
	Duration       string
	LastFrom       time.Time
	NextPage       string
	SkipCollection bool
	SkipVideos     bool
	RedditFilter   reddit.PostFilter
}

type Source interface {
	ScrapePosts(context.Context, string, ScrapeOpts) (chan Post, error)
	DownloadItem(context.Context, Item) ([]byte, error)
}
