package sources

import (
	"time"

	"github.com/shivamhw/content-pirate/commons"
	"github.com/shivamhw/content-pirate/pkg/reddit"
)


type SourceType string

const (
     SOURCE_TYPE_REDDIT SourceType = "SOURCE_TYPE_REDDIT"
     SOURCE_TYPE_TELEGRAM SourceType = "SOURCE_TYPE_TELEGRAM"
)

type Post struct {
	MediaType commons.MediaType
	SrcLink   string
	Title     string
	Id        string
	SourceAc  string
	Ext       string
	FileName  string
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