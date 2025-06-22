package scrapper

import (
	"github.com/shivamhw/content-pirate/sources"
	"github.com/shivamhw/content-pirate/store"
)

type SourceType string

const (
	REDDIT   SourceType = "REDDIT"
	TELEGRAM SourceType = "TELEGRAM"
)

type Job struct {
	SrcAc       string
	SrcId       string
	Dst         store.DstPath
	Opts        JobOpts
	SourceStore SourceType
}

type JobOpts = sources.ScrapeOpts
// type JobOpts struct {
// 	Limit          int
// 	Page           int
// 	Last           string
// 	Duration       string
// 	LastFrom       time.Time
// 	NextPage       string
// 	SkipCollection bool
// 	SkipVideos     bool
// }



