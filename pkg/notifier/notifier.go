package notifier

import (
	"context"

	"github.com/shivamhw/content-pirate/commons"
)

type NotifierEvent string

const (
	NOTIFY_ITEM_SAVED NotifierEvent = "item_saved"
)

type Notifier interface {
	Notify(ctx context.Context, item *commons.Item, event NotifierEvent) error
}