package sources

import (
	"context"
	"fmt"
	"log"
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
	CfgPath       string
	ctx           context.Context
	duration      string
	skipCollection bool
}

func NewRedditClient(opts RedditClientOpts) *RedditClient {
	redditClient := &RedditClient{
		aCfg: &authCfg{},
		ctx:  opts.ctx,
		opts: &opts,
	}
	if opts.CfgPath == "" {
		log.Printf("no reddit config passed using default client")
		return redditClient
	}
	GetCfgFromJson(opts.CfgPath, redditClient.aCfg)
	// create auth
	credentials := reddit.Credentials{
		ID:       redditClient.aCfg.ID,
		Secret:   redditClient.aCfg.Secret,
		Username: redditClient.aCfg.Username,
		Password: redditClient.aCfg.Password,
	}
	c, err := reddit.NewClient(credentials)
	if err != nil {
		log.Printf("err creating client %s, using default client", err)
		c = reddit.DefaultClient()
	}
	redditClient.client = c
	return redditClient
}

func (r *RedditClient) Scrape(subreddit string, p chan<- Post) {
	rposts, err := r.GetTopPosts(subreddit)
	if err != nil {
		log.Printf("scrapping subreddit %s failed with %s", subreddit, err)
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
			log.Print("found gallery ", post.URL)
			for _, item := range post.GalleryData.Items {
				link := fmt.Sprintf("https://i.redd.it/%s.%s", item.MediaID, GetMIME(post.MediaMetadata[item.MediaID].MIME))
				log.Printf("created link %s for gal %s %s", link, post.Title, item.MediaID)
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
					if !r.opts.skipCollection {
						log.Println("not downloading full collection")
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
			Time: r.opts.duration,
		})
		if err != nil {
			if strings.Contains(err.Error(), "429") {
				log.Printf("HIT rate limit wait 2 sec")
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
