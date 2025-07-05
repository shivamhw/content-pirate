package scrapper

import (
	"github.com/shivamhw/content-pirate/sources"
	"github.com/shivamhw/content-pirate/store"
)


type Job struct {
	SrcAc       string
	SrcId       string
	Dst         []store.DstPath	`json:"-"`
	Opts        JobOpts
}

type JobOpts = sources.ScrapeOpts
