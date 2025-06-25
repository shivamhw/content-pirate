package sources

import (
	"context"
	"fmt"
	"io"
	"net/http"
	"strings"

	"github.com/shivamhw/content-pirate/commons"
	. "github.com/shivamhw/content-pirate/pkg/log"
	"github.com/shivamhw/content-pirate/pkg/reddit"
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
		CfgPath: opts.CfgPath,
	})
	if err != nil {
		return nil, err
	}
	return &RedditStore{
		client: c,
		opts:   opts,
	}, nil
}

func (r *RedditStore) ScrapePosts(_ context.Context, subreddit string, opts ScrapeOpts) (p chan Post, err error) {
	p = make(chan Post, 5)
	cnt := 0
	rOpts := reddit.ListOptions{
		Limit:    opts.Limit,
		Page:     opts.Page,
		NextPage: opts.NextPage,
		Filter:   opts.RedditFilter,
		Duration: opts.Duration,
	}
	rposts, err := r.client.GetPosts(subreddit, rOpts)
	if err != nil {
		Logger.Error("scrapping subreddit failed ", "subreddit", subreddit, "error", err)
	}
	go func() {
		defer func() {
			close(p)
			Logger.Info("scrapping post completed.", "scraped posts", cnt)
		}()
		posts := r.convertToPosts(rposts, subreddit, opts)
		for _, post := range posts {
			p <- post
			cnt++
		}
	}()
	return p, nil
}

func (r *RedditStore) convertToPosts(rposts []*reddit.Post, subreddit string, opts ScrapeOpts) (posts []Post) {
	for _, post := range rposts {
		// if gallary link
		if strings.Contains(post.URL, "/gallery/") {
			Logger.Debug("found gallery", "url", post.URL)
			for _, item := range post.GalleryData.Items {
				link := fmt.Sprintf("https://i.redd.it/%s.%s", item.MediaID, commons.GetMIME(post.MediaMetadata[item.MediaID].MIME))
				Logger.Debug("created", "link", link, "post title", post.Title, "mediaId", item.MediaID)
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

func (r *RedditStore) DownloadItem(ctx context.Context, i Item) ([]byte, error) {
	req, err := http.NewRequestWithContext(ctx, http.MethodGet, i.Src, nil)
	if err != nil {
		return nil, err
	}
	resp, err := http.DefaultClient.Do(req)
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
