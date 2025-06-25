package sources

import (
	"context"

	"github.com/shivamhw/content-pirate/commons"
	"github.com/shivamhw/content-pirate/store"
)

type ItemStatus string

const (
	FAILED  ItemStatus = "FAILED"
	SUCCESS ItemStatus = "SUCCESS"
	STARTED ItemStatus = "STARTED"
)

type Item struct {
	Id       string
	TaskId   string
	Src      string
	FileName string
	Type     commons.MediaType
	DstStore store.Store `json:"-"`
	Dst      string
	Status   ItemStatus
	SourceAc string
	Ext      string
	Title    string
	Ctx      context.Context `json:"-"`
}

type ItemUpdateOpts struct {
	Dst      string
	Status   ItemStatus
	FileName string
}
