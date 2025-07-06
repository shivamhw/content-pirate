package telegram_cmd

import (
	"fmt"
	"time"

	"github.com/shivamhw/content-pirate/pkg/scrapper"
	"github.com/shivamhw/content-pirate/sources"
	"github.com/shivamhw/content-pirate/store"
	"github.com/spf13/cobra"
)

var (
	sCfg       scrapper.ScrapeCfg
	scrapeOpts sources.ScrapeOpts
	dst        store.TelegramDstPath
	ids        []string
)

func scrapeCmd() *cobra.Command {
	var jIds []string
	var timeDelta int
	var waitTime int
	cmd := &cobra.Command{
		Use:   "scrape",
		Long:  "Scrapes chats for videos and imgs",
		Short: "scrapes chats",
		RunE: func(cmd *cobra.Command, args []string) error {
			ids = UniqueStrings(ids)
			sCfg.SourceType = sources.SOURCE_TYPE_TELEGRAM
			s, err := scrapper.NewScrapper(&sCfg)
			if err != nil {
				return err
			}
			go s.Start()
			scrapeOpts.LastFrom = time.Now().Add(time.Duration(-timeDelta) * time.Minute)
			dst.PhoneNumber = sCfg.PhoneNumber
			count := 0
			for {
				fmt.Println("sent msg", count)
				for _, i := range ids {
					j := scrapper.Job{
						SrcAc: i,
						Dst:   []store.DstPath{dst},
						Opts:  scrapeOpts,
					}

					id, err := s.SubmitJob(j)
					if err != nil {
						return err
					}
					jIds = append(jIds, id)
				}
				fmt.Println("waiting...")
				time.Sleep(time.Duration(waitTime) * time.Minute)
				count = 0
				for _, i := range jIds {
					j, _ := s.GetJob(i)
					fmt.Printf("%s\t%d\n", j.J.SrcAc, len(j.I))
					count += len(j.I)
				}
			}
		},
	}
	cmd.Flags().IntVar(&scrapeOpts.Limit, "limit", 25, "limit")
	cmd.Flags().StringSliceVar(&ids, "source", []string{}, "source channel ids")
	cmd.Flags().IntVar(&sCfg.ImgWorkers, "img-worker", 1, "nof img proccesing worker")
	cmd.Flags().IntVar(&sCfg.VidWorkers, "vid-worker", 1, "nof vid proccesing worker")
	cmd.Flags().Int64Var(&sCfg.TimeOut, "time-out", 60, "timeout in seconds")
	cmd.Flags().IntVar(&sCfg.TopicWorkers, "reddit-worker", 15, "nof reddit proccesing worker")
	cmd.Flags().StringVar(&sCfg.PhoneNumber, "phone", "", "phone nm for telegram")
	cmd.Flags().IntVar(&timeDelta, "last", 60, "last msgs from x minutes")
	cmd.Flags().IntVar(&waitTime, "wait", 1, "wait in x minutes")
	cmd.Flags().IntVar(&dst.ChatId, "dst", 0, "dst channel id")
	return cmd
}

func UniqueStrings(input []string) []string {
	seen := make(map[string]struct{})
	var result []string

	for _, val := range input {
		if _, exists := seen[val]; !exists {
			seen[val] = struct{}{}
			result = append(result, val)
		}
	}
	return result
}
