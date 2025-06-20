package reddit

import (
	"github.com/vartanbeno/go-reddit/v2/reddit"
	"github.com/jinzhu/copier"

)

type Post struct {
	ID      string     `json:"id,omitempty"`
	FullID  string     `json:"name,omitempty"`
	Created *reddit.Timestamp `json:"created_utc,omitempty"`
	Edited  *reddit.Timestamp `json:"edited,omitempty"`

	Permalink string `json:"permalink,omitempty"`
	URL       string `json:"url,omitempty"`

	Title string `json:"title,omitempty"`
	Body  string `json:"selftext,omitempty"`

	// Indicates if you've upvoted/downvoted (true/false).
	// If neither, it will be nil.
	Likes *bool `json:"likes"`

	GalleryData reddit.GalleryData `json:"gallery_data,omitempty"`

	MediaMetadata reddit.MediaMetadata `json:"media_metadata,omitempty"`

	Media reddit.Media `json:"media,omitempty"`

	Score            int     `json:"score"`
	UpvoteRatio      float32 `json:"upvote_ratio"`
	NumberOfComments int     `json:"num_comments"`

	SubredditName         string `json:"subreddit,omitempty"`
	SubredditNamePrefixed string `json:"subreddit_name_prefixed,omitempty"`
	SubredditID           string `json:"subreddit_id,omitempty"`
	SubredditSubscribers  int    `json:"subreddit_subscribers"`

	Author   string `json:"author,omitempty"`
	AuthorID string `json:"author_fullname,omitempty"`

	Spoiler    bool `json:"spoiler"`
	Locked     bool `json:"locked"`
	NSFW       bool `json:"over_18"`
	IsSelfPost bool `json:"is_self"`
	Saved      bool `json:"saved"`
	Stickied   bool `json:"stickied"`
}

func ConvertFrom(v reddit.Post) *Post {
	var p Post
	copier.Copy(&p, &v)
	return &p
}