package sources

import (
	"context"
	"github.com/shivamhw/content-pirate/commons"
)



type Source interface {
	ScrapePosts(context.Context, string, ScrapeOpts) (chan Post, error)
	DownloadItem(context.Context, *commons.Item) (error)
}
