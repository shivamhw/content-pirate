package sources

import (
	"context"
	"fmt"
	"io"
	"net/http"
	"strings"

	"github.com/shivamhw/reddit-pirate/commons"
	. "github.com/shivamhw/reddit-pirate/pkg/log"
	"github.com/shivamhw/reddit-pirate/pkg/reddit"
)

const (
	DEFAULT_LIMIT = 100
)

type RedditStore struct {
	client *reddit.RedditClient
	opts  *RedditStoreOpts
}

type RedditStoreOpts struct {
	reddit.RedditClientOpts
}

func NewRedditStore(ctx context.Context, opts *RedditStoreOpts) (*RedditStore, error) {
	c, err := reddit.NewRedditClient(ctx, reddit.RedditClientOpts{
		CfgPath: opts.CfgPath,
		Duration: opts.Duration,
		SkipCollection: opts.SkipCollection,
	})
	if err != nil {
		return nil, err
	}
	return &RedditStore{
		client: c,
		opts: opts,
	}, nil
}

func (r *RedditStore) ScrapePosts(subreddit string, p chan<- commons.Post) {
	rposts, err := r.client.GetTopPosts(subreddit, DEFAULT_LIMIT)
	if err != nil {
		Logger.Error("scrapping subreddit failed ","subreddit" ,subreddit,"error", err)
	}
	posts := r.convertToPosts(rposts, subreddit)
	for _, post := range posts {
		p <- post
	}
}

func (r *RedditStore) convertToPosts(rposts []*reddit.Post, subreddit string) (posts []commons.Post) {
	for _, post := range rposts {
		// if gallary link
		if strings.Contains(post.URL, "/gallery/") {
			Logger.Info("found gallery","url", post.URL)
			for _, item := range post.GalleryData.Items {
				link := fmt.Sprintf("https://i.redd.it/%s.%s", item.MediaID, commons.GetMIME(post.MediaMetadata[item.MediaID].MIME))
				Logger.Info("created","link", link, "post title", post.Title,"mediaId", item.MediaID)
				if commons.IsImgLink(link) {
					post := commons.Post{
						Id:        post.ID,
						Title:     fmt.Sprintf("%s_GAL_%s", post.Title, item.MediaID[:len(item.MediaID)-3]),
						MediaType: commons.IMG_TYPE,
						Ext:       commons.GetMIME(post.MediaMetadata[item.MediaID].MIME),
						SrcLink:   link,
						SourceAc:  subreddit,
					}
					posts = append(posts, post)
					if !r.opts.SkipCollection {
						Logger.Info("not downloading full collection")
						break
					}
				}
			}
			continue
		}
		// if single img post
		if commons.IsImgLink(post.URL) {
			p := commons.Post{
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
		if post.Media.RedditVideo.FallbackURL != "" {
			p := commons.Post{
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

func (r *RedditStore) DownloadJob(j commons.Job) ([]byte,error) {
	resp, err := http.Get(j.Src)
	if err != nil || resp.StatusCode != http.StatusOK {
		return nil, fmt.Errorf("failed to download %s because %s code", j.Src, err)
	}
	defer resp.Body.Close()
	data, err := io.ReadAll(resp.Body)
	if err != nil {
		return nil, fmt.Errorf("error downloading job %s err %s", j.Src, err.Error())
	}
	return data, nil
}