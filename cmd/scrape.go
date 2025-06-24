package cmd

import (
	"fmt"
	"strings"

	"github.com/shivamhw/content-pirate/pkg/reddit"
	"github.com/shivamhw/content-pirate/pkg/scrapper"
	"github.com/shivamhw/content-pirate/sources"
	"github.com/shivamhw/content-pirate/store"
	"github.com/spf13/cobra"
)

var (
	sCfg       scrapper.ScrapeCfg
	scrapeOpts sources.ScrapeOpts
	dst        store.DstPath
	ids        []string
)

func scrapeCmd() *cobra.Command {
	var jIds []string
	var filter string
	cmd := &cobra.Command{
		Use:   "scrape",
		Long:  "Scrapes subreddit for videos and imgs",
		Short: "scrapes subreddit",
		RunE: func(cmd *cobra.Command, args []string) error {
			f1, err := sanitizeFilter(&filter)
			if err != nil {
				return err
			}
			scrapeOpts.RedditFilter = f1
			s, err := scrapper.NewScrapper(&sCfg)
			if err != nil {
				return err
			}
			go s.Start()
			for _, i := range ids {
				j := scrapper.Job{
					SrcAc:       i,
					Dst:         dst,
					Opts:        scrapeOpts,
					SourceStore: scrapper.REDDIT,
				}

				id, err := s.SubmitJob(j)
				if err != nil {
					return err
				}
				jIds = append(jIds, id)
			}
			for _, i := range jIds {
				s.WaitOnId(i)
			}
			for _, i := range jIds {
				j, _ := s.GetJob(i)
				for _, f := range j.I {
					fmt.Printf("%s\t%s\n", f.SourceAc, f.Dst)
				}
			}
			return err
		},
	}
	cmd.Flags().StringVar(&dst.BasePath, "dir", "./download", "dst folder for downloads")
	cmd.Flags().StringVar(&dst.ImgPath, "img-dir", "imgs", "dst folder for imgs")
	cmd.Flags().StringVar(&dst.VidPath, "vid-dir", "vids", "dst folder for vids")
	cmd.Flags().StringVar(&sCfg.AuthCfg, "auth", "./reddit.json", "auth config for reddit")
	cmd.Flags().StringVar(&scrapeOpts.Duration, "duration", "day", "duration")
	cmd.Flags().IntVar(&scrapeOpts.Limit, "limit", 25, "limit")
	cmd.Flags().StringSliceVar(&ids, "source", []string{}, "source channel ids")
	cmd.Flags().BoolVar(&scrapeOpts.SkipVideos, "skip-vid", true, "skip video download")
	cmd.Flags().BoolVar(&dst.CombineDir, "combine", false, "combine folders")
	cmd.Flags().BoolVar(&scrapeOpts.SkipCollection, "skip-collection", false, "download full collection")
	cmd.Flags().BoolVar(&dst.CleanOnStart, "cleanOnStart", false, "clean folders")
	cmd.Flags().IntVar(&sCfg.ImgWorkers, "img-worker", 10, "nof img proccesing worker")
	cmd.Flags().IntVar(&sCfg.VidWorkers, "vid-worker", 5, "nof vid proccesing worker")
	cmd.Flags().IntVar(&sCfg.TopicWorkers, "reddit-worker", 15, "nof reddit proccesing worker")
	cmd.Flags().StringVar(&filter, "filter", "TOP", "filter: NEW, HOT, TOP")
	return cmd
}


func sanitizeFilter(f *string) (reddit.PostFilter, error) {
	filter := strings.ToUpper(fmt.Sprintf("REDDIT_%s", *f))
	f1 := reddit.PostFilter(filter)
	if f1 != reddit.REDDIT_HOT && f1 != reddit.REDDIT_NEW && f1 != reddit.REDDIT_TOP {
		return reddit.PostFilter(""), fmt.Errorf("unknown filter for reddit %s", f1)
	}
	return f1, nil
}