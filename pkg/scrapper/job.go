package scrapper

import (
	"github.com/shivamhw/content-pirate/sources"
	"github.com/shivamhw/content-pirate/store"
)


type Job struct {
	SrcAc       string
	SrcId       string
	Dst         []store.DstPath
	Opts        JobOpts
	SourceType sources.SourceType
}

type JobOpts = sources.ScrapeOpts
