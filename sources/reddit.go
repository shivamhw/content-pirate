package sources

import (
	"context"
	"fmt"
	"io"
	log "log/slog"
	"net/http"
	"os"
	"strings"
	"time"

	. "github.com/shivamhw/reddit-pirate/commons"
	"github.com/vartanbeno/go-reddit/v2/reddit"
)

type RedditClient struct {
	client *reddit.Client
	aCfg   *authCfg
	ctx    context.Context
	opts   *RedditClientOpts
}

type authCfg struct {
	ID       string `json:"id"`
	Secret   string `json:"secret"`
	Username string `json:"username"`
	Password string `json:"password"`
}

type RedditClientOpts struct {
	CfgPath        string
	Ctx            context.Context
	Duration       string
	SkipCollection bool
}

func NewRedditClient(opts RedditClientOpts) *RedditClient {
	redditClient := &RedditClient{
		aCfg: &authCfg{},
		ctx:  opts.Ctx,
		client: reddit.DefaultClient(),
		opts: &opts,
	}
	err := ReadFromJson(opts.CfgPath, redditClient.aCfg)
	if os.IsNotExist(err) {
		log.Warn("file does not exists")
		opts.CfgPath = ""
	}
	if opts.CfgPath == "" {
		log.Warn("no reddit config passed using default client")
		return redditClient
	}
	// create auth
	credentials := reddit.Credentials{
		ID:       redditClient.aCfg.ID,
		Secret:   redditClient.aCfg.Secret,
		Username: redditClient.aCfg.Username,
		Password: redditClient.aCfg.Password,
	}
	c, err := reddit.NewClient(credentials)
	if err != nil {
		log.Error("err creating client, using default client","error", err)
		return redditClient
	}
	redditClient.client = c
	return redditClient
}

func (r *RedditClient) ScrapePosts(subreddit string, p chan<- Post) {
	rposts, err := r.GetTopPosts(subreddit)
	if err != nil {
		log.Error("scrapping subreddit failed ","subreddit" ,subreddit,"error", err)
	}
	posts := r.convertToPosts(rposts, subreddit)
	for _, post := range posts {
		p <- post
	}
}

func (r *RedditClient) convertToPosts(rposts []*reddit.Post, subreddit string) (posts []Post) {
	for _, post := range rposts {
		// if gallary link
		if strings.Contains(post.URL, "/gallery/") {
			log.Info("found gallery","url", post.URL)
			for _, item := range post.GalleryData.Items {
				link := fmt.Sprintf("https://i.redd.it/%s.%s", item.MediaID, GetMIME(post.MediaMetadata[item.MediaID].MIME))
				log.Info("created","link", link, "post title", post.Title,"mediaId", item.MediaID)
				if IsImgLink(link) {
					post := Post{
						Id:        post.ID,
						Title:     fmt.Sprintf("%s_GAL_%s", post.Title, item.MediaID[:len(item.MediaID)-3]),
						MediaType: IMG_TYPE,
						Ext:       GetMIME(post.MediaMetadata[item.MediaID].MIME),
						SrcLink:   link,
						SourceAc:  subreddit,
					}
					posts = append(posts, post)
					if !r.opts.SkipCollection {
						log.Info("not downloading full collection")
						break
					}
				}
			}
			continue
		}
		// if single img post
		if IsImgLink(post.URL) {
			p := Post{
				Id:        post.ID,
				Title:     post.Title,
				SrcLink:   post.URL,
				SourceAc:  subreddit,
				Ext:       GetExtFromLink(post.URL),
				MediaType: IMG_TYPE,
			}
			posts = append(posts, p)
			continue
		}
		if post.Media.RedditVideo.FallbackURL != "" {
			p := Post{
				Id:        post.ID,
				Title:     post.Title,
				MediaType: VID_TYPE,
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

func (r *RedditClient) GetTopPosts(subreddit string) ([]*reddit.Post, error) {
	var final_posts []*reddit.Post
	nextToken := ""
	for {
		posts, resp, err := r.client.Subreddit.TopPosts(r.ctx, subreddit, &reddit.ListPostOptions{
			ListOptions: reddit.ListOptions{
				Limit: 100,
				After: nextToken,
			},
			Time: r.opts.Duration,
		})
		if err != nil {
			if strings.Contains(err.Error(), "429") {
				log.Warn("HIT rate limit wait 2 sec")
				time.Sleep(2 * time.Second)
				continue
			} else {
				return nil, err
			}
		}
		final_posts = append(final_posts, posts...)
		nextToken = resp.After
		if nextToken == "" {
			break
		}
	}
	return final_posts, nil
}

func (r *RedditClient) DownloadJob(j Job) ([]byte,error) {
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