package sources

import (
	"context"
	"log"
	"strings"
	"sync"
	"time"

	. "github.com/shivamhw/reddit-pirate/commons"
	"github.com/vartanbeno/go-reddit/v2/reddit"
)

type RedditClient struct {
	client   *reddit.Client
	Cfg      *authCfg
	ctx      context.Context
	duration string
	workerQ  chan reddit.Post
	threads  int
}

type authCfg struct {
	ID       string `json:"id"`
	Secret   string `json:"secret"`
	Username string `json:"username"`
	Password string `json:"password"`
}

type RedditClientOpts struct {
	CfgPath  string
	ctx      context.Context
	duration string
	threads  int
}

func NewRedditClient(opts RedditClientOpts) *RedditClient {
	redditClient := &RedditClient{
		Cfg:      &authCfg{},
		ctx:      opts.ctx,
		duration: opts.duration,
		threads:  opts.threads,
	}
	if opts.CfgPath == "" {
		log.Printf("no reddit config passed using default client")
		return redditClient
	}
	GetCfgFromJson(opts.CfgPath, redditClient.Cfg)
	// create auth
	credentials := reddit.Credentials{
		ID:       redditClient.Cfg.ID,
		Secret:   redditClient.Cfg.Secret,
		Username: redditClient.Cfg.Username,
		Password: redditClient.Cfg.Password,
	}
	c, err := reddit.NewClient(credentials)
	if err != nil {
		log.Printf("err creating client %s, using default client", err)
		c = reddit.DefaultClient()
	}
	redditClient.client = c
	return redditClient
}

func (r *RedditClient) Emit(subreddits []string) <-chan Post {
	p := make(chan Post)
	s := make(chan string)
	var wg sync.WaitGroup
	// emit to central q
	for i := 0; i < r.threads; i++ {
		wg.Add(1)
		go r.scrapperWorker(s, &wg)
	}
	go func() {
		for _, sub := range subreddits {
			log.Printf("adding subreddit %s to q", sub)
			s <- sub
		}
		close(s)
		wg.Wait()
		close(r.workerQ)
	}()
	
	// read from central q and write to extrenal q

	

	return p
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
			Time: r.duration,
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

func (r *RedditClient) scrapperWorker(subQ <-chan string, wg *sync.WaitGroup) {
	//scrape all the post
	defer wg.Done()
	for subreddit := range subQ {
		log.Printf("started proccessing subreddit %s", subreddit)
		posts, err := r.GetTopPosts(subreddit)
		// write to q
		if err != nil {
			log.Printf("unable to scrape subreddit %s", subreddit)
		}
		for _, post := range posts {
			r.workerQ <- *post
		}
	}
}
