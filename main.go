package main

import (
	"context"
	"fmt"
    "github.com/shivamhw/reddit-pirate/cmd"
	"github.com/shivamhw/reddit-pirate/pkg/reddit"
)

  

func main() {
	cmd.Execute()
	// redditTest()
}

func redditTest(){
   r, _ := reddit.NewRedditClient(context.Background(), reddit.RedditClientOpts{})
   res, _ := r.SearchSubreddits("unix", 20)
   for _, sub := range res {
	fmt.Printf("thiis si sub %s\n", sub.Name)
   }
}
