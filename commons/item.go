package commons

import (
	"context"
)

type ItemStatus string

const (
	FAILED  ItemStatus = "FAILED"
	SUCCESS ItemStatus = "SUCCESS"
	STARTED ItemStatus = "STARTED"
)

type Item struct {
	Id       string
	Src      string
	FileName string
	Type     MediaType
	Dst      string
	DstId    string
	Status   ItemStatus
	SourceAc string
	Size     int64
	Ext      string
	Title    string
	Ctx      context.Context `json:"-"`
	Data     []byte   `json:"-"`
}

type ItemUpdateOpts struct {
	Dst      string
	Status   ItemStatus
	FileName string
}
