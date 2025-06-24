package reddit

import (
	"context"
	"os"
	"strings"
	"time"

	"github.com/shivamhw/content-pirate/commons"
	. "github.com/shivamhw/content-pirate/pkg/log"
	"github.com/vartanbeno/go-reddit/v2/reddit"
)

type PostFilter string

const (
	REDDIT_TOP    PostFilter = "REDDIT_TOP"
	REDDIT_HOT    PostFilter = "REDDIT_HOT"
	REDDIT_NEW    PostFilter = "REDDIT_NEW"
	DEFAULT_LIMIT            = 10
)

type RedditClient struct {
	Client *reddit.Client
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
	CfgPath string
}

type ListOptions struct {
	Limit    int
	Page     int
	NextPage string
	Duration string // accept hour, day
	Filter   PostFilter
}

func NewRedditClient(ctx context.Context, opts RedditClientOpts) (*RedditClient, error) {
	redditClient := &RedditClient{
		aCfg:   &authCfg{},
		ctx:    ctx,
		Client: reddit.DefaultClient(),
		opts:   &opts,
	}
	err := commons.ReadFromJson(opts.CfgPath, redditClient.aCfg)
	if os.IsNotExist(err) {
		Logger.Warn("file does not exists", "file", opts.CfgPath)
		opts.CfgPath = ""
	}
	if opts.CfgPath == "" {
		Logger.Warn("no reddit config passed using default client")
		return redditClient, nil
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
		Logger.Error("err creating client, using default client", "error", err)
		return redditClient, err
	}
	redditClient.Client = c
	return redditClient, nil
}

func (r *RedditClient) GetPosts(subreddit string, opts ListOptions) ([]*Post, error) {
	var final_posts []*Post
	var posts []*reddit.Post
	var resp *reddit.Response
	var err error
	opts.sanitize()
	nextToken := opts.NextPage
	Logger.Info("scarpping reddit", "filter", opts.Filter, "limit", opts.Limit)
	for {
		page := min(opts.Limit, 25)
		opts.Limit -= page

		switch opts.Filter {
		case REDDIT_TOP:
			posts, resp, err = r.Client.Subreddit.TopPosts(r.ctx, subreddit, &reddit.ListPostOptions{
				ListOptions: reddit.ListOptions{
					Limit: page,
					After: nextToken,
				},
				Time: opts.Duration,
			})
		case REDDIT_HOT:
			posts, resp, err = r.Client.Subreddit.HotPosts(r.ctx, subreddit, &reddit.ListOptions{
				Limit: page,
				After: nextToken,
			})
		case REDDIT_NEW:
			posts, resp, err = r.Client.Subreddit.NewPosts(r.ctx, subreddit, &reddit.ListOptions{
				Limit: page,
				After: nextToken,
			})
		}

		if err != nil {
			if strings.Contains(err.Error(), "429") {
				Logger.Warn("HIT rate limit wait 2 sec")
				time.Sleep(2 * time.Second)
				continue
			} else {
				return nil, err
			}
		}
		
		for _, p := range posts {
			final_posts = append(final_posts, ConvertFrom(*p))
		}
		nextToken = resp.After
		if nextToken == "" || opts.Limit <= 0 {
			break
		}
	}
	return final_posts, nil
}

func (r *RedditClient) GetSubscribedSubreddits(limit int) ([]*reddit.Subreddit, error) {
	nextToken := ""
	var err error
	var results []*reddit.Subreddit
	for {
		subs, resp, err := r.Client.Subreddit.Subscribed(r.ctx, &reddit.ListSubredditOptions{
			ListOptions: reddit.ListOptions{
				Limit: limit,
				After: nextToken,
			},
		})
		if err != nil {
			Logger.Error("failed getting subcribed subreddit list", "err", err)
		}
		nextToken = resp.After
		results = append(results, subs...)
		if nextToken == "" {
			break
		}
	}
	return results, err
}

func (r *RedditClient) SearchSubreddits(q string, limit int) ([]*reddit.Subreddit, error) {
	nextToken := ""
	var err error
	var results []*reddit.Subreddit
	for {
		page := min(limit, 25)
		limit -= page
		subs, resp, err := r.Client.Subreddit.Search(r.ctx, q, &reddit.ListSubredditOptions{
			ListOptions: reddit.ListOptions{
				Limit: page,
				After: nextToken,
			},
		})
		if err != nil {
			Logger.Error("failed getting search subreddit list", "err", err)
		}
		nextToken = resp.After
		results = append(results, subs...)
		if nextToken == "" || limit <= 0 {
			break
		}
	}
	return results, err
}

func (l *ListOptions) sanitize() {
	if l.Limit == 0 {
		l.Limit = DEFAULT_LIMIT
	}
	if l.Duration == "" {
		l.Duration = "day"
	}
	if l.Filter == "" {
		l.Filter = REDDIT_TOP
	}
}
