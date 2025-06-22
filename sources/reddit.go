package sources

import (
	"context"
	"fmt"
	"io"
	"net/http"
	"strings"
	"sync/atomic"

	"github.com/shivamhw/content-pirate/commons"
	. "github.com/shivamhw/content-pirate/pkg/log"
	"github.com/shivamhw/content-pirate/pkg/reddit"
)

const (
	DEFAULT_LIMIT = 10
)

type RedditStore struct {
	client *reddit.RedditClient
	opts   *RedditStoreOpts
}

type RedditStoreOpts struct {
	reddit.RedditClientOpts
}

func NewRedditStore(ctx context.Context, opts *RedditStoreOpts) (*RedditStore, error) {
	c, err := reddit.NewRedditClient(ctx, reddit.RedditClientOpts{
		CfgPath:        opts.CfgPath,
	})
	if err != nil {
		return nil, err
	}
	return &RedditStore{
		client: c,
		opts:   opts,
	}, nil
}

func (r *RedditStore) ScrapePosts(subreddit string, opts ScrapeOpts) (p chan Post, err error) {
	p = make(chan Post, 5)
	var count int64
	if opts.Limit <= 0 {
		opts.Limit = DEFAULT_LIMIT
	}
	rposts, err := r.client.GetTopPosts(subreddit, reddit.ListOptions{
		Limit:    opts.Limit,
		Page:     opts.Page,
		NextPage: opts.NextPage,
		Duration: opts.Duration,
	})
	if err != nil {
		Logger.Error("scrapping subreddit failed ", "subreddit", subreddit, "error", err)
	}
	go func(c *int64) {
		defer func(){
		close(p)
		Logger.Info("scrapping post completed.","scraped posts", *c)
		}()
		posts := r.convertToPosts(rposts, subreddit, opts)
		for _, post := range posts {
			p <- post
			atomic.AddInt64(c, 1)
		}
	}(&count)
	return p, nil
}

func (r *RedditStore) convertToPosts(rposts []*reddit.Post, subreddit string, opts ScrapeOpts) (posts []Post) {
	for _, post := range rposts {
		// if gallary link
		if strings.Contains(post.URL, "/gallery/") {
			Logger.Info("found gallery", "url", post.URL)
			for _, item := range post.GalleryData.Items {
				link := fmt.Sprintf("https://i.redd.it/%s.%s", item.MediaID, commons.GetMIME(post.MediaMetadata[item.MediaID].MIME))
				Logger.Info("created", "link", link, "post title", post.Title, "mediaId", item.MediaID)
				if commons.IsImgLink(link) {
					post := Post{
						Id:        fmt.Sprintf("%d", item.ID),
						Title:     post.Title, //fmt.Sprintf("%s_GAL_%s", post.Title, item.MediaID[:len(item.MediaID)-3]),
						MediaType: commons.IMG_TYPE,
						Ext:       commons.GetMIME(post.MediaMetadata[item.MediaID].MIME),
						SrcLink:   link,
						SourceAc:  subreddit,
					}
					posts = append(posts, post)
					if opts.SkipCollection {
						Logger.Info("not downloading full collection")
						break
					}
				}
			}
			continue
		}
		// if single img post
		if commons.IsImgLink(post.URL) {
			p := Post{
				Id:        post.ID,
				Title:     post.Title,
				SrcLink:   post.URL,
				SourceAc:  subreddit,
				Ext:       commons.GetExtFromLink(post.URL),
				MediaType: commons.IMG_TYPE,
			}
			posts = append(posts, p)
			continue
		}
		if !opts.SkipVideos && post.Media.RedditVideo.FallbackURL != "" {
			p := Post{
				Id:        post.ID,
				Title:     post.Title,
				MediaType: commons.VID_TYPE,
				SrcLink:   post.Media.RedditVideo.FallbackURL,
				Ext:       "mp4",
				SourceAc:  subreddit,
			}
			posts = append(posts, p)
			continue
		}
	}
	return
}

func (r *RedditStore) DownloadItem(i Item) ([]byte, error) {
	resp, err := http.Get(i.Src)
	if err != nil || resp.StatusCode != http.StatusOK {
		return nil, fmt.Errorf("failed to download %s because %s code", i.Src, err)
	}
	defer resp.Body.Close()
	data, err := io.ReadAll(resp.Body)
	if err != nil {
		return nil, fmt.Errorf("error downloading job %s err %s", i.Src, err.Error())
	}
	return data, nil
}
