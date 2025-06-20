package cmd

import (
	"context"
	"fmt"
	log "log/slog"
	"path/filepath"
	"sync"
	"sync/atomic"

	. "github.com/shivamhw/content-pirate/commons"
	"github.com/shivamhw/content-pirate/pkg/reddit"
	"github.com/shivamhw/content-pirate/sources"
	"github.com/shivamhw/content-pirate/store"
	"github.com/spf13/cobra"
)

type Scrapper struct {
	SourceStore sources.Source
	DstStore    store.Store
	SourceAc    []string
	sCfg        *ScrapeCfg
	ctx         context.Context
	dstPath     *DstPath
}

type AuthCfg struct {
	ID       string `json:"id"`
	Secret   string `json:"secret"`
	Username string `json:"username"`
	Password string `json:"password"`
}

type ScrapeCfg struct {
	dstDir         string
	vidDir         string
	imgDir         string
	authCfg        string
	subreddits     string
	postId         string
	skipVideo      bool
	cleanOnStart   bool
	combineDir     bool
	skipCollection bool
	imgWorker      int
	vidWorker      int
	redWorker      int
	scrapeOpts  *sources.ScrapeOpts
	sourceIds      []string
}

type Mediums struct {
	subq  chan string
	postq chan Post
	imgq  chan Job
	vidq  chan Job
	swg   sync.WaitGroup
}

var (
	sCfg       ScrapeCfg
	imgCounter int64
	vidCounter int64
	scrapeOpts sources.ScrapeOpts
)

func scrapeCmd() *cobra.Command {
	cmd := &cobra.Command{
		Use:   "scrape",
		Long:  "Scrapes subreddit for videos and imgs",
		Short: "scrapes subreddit",
		RunE: func(cmd *cobra.Command, args []string) error {
			scr, err := NewScrapper(&sCfg)
			if err != nil {
				return err
			}
			return scr.Start()
		},
	}
	cmd.Flags().StringVar(&sCfg.dstDir, "dir", "./download", "dst folder for downloads")
	cmd.Flags().StringVar(&sCfg.imgDir, "img-dir", "imgs", "dst folder for imgs")
	cmd.Flags().StringVar(&sCfg.vidDir, "vid-dir", "vids", "dst folder for vids")
	cmd.Flags().StringVar(&sCfg.subreddits, "subs", "./subreddits.json", "list of subreddits")
	cmd.Flags().StringVar(&sCfg.authCfg, "auth", "./reddit.json", "auth config for reddit")
	cmd.Flags().StringVar(&sCfg.postId, "post-id", "", "post id")
	cmd.Flags().StringVar(&scrapeOpts.Duration, "duration", "day", "duration")
	cmd.Flags().IntVar(&scrapeOpts.Limit, "limit", 25, "limit")
	cmd.Flags().StringSliceVar(&sCfg.sourceIds, "source", []string{}, "source channel ids")
	cmd.Flags().BoolVar(&sCfg.skipVideo, "skip-vid", true, "skip video download")
	cmd.Flags().BoolVar(&sCfg.combineDir, "combine", true, "combine folders")
	cmd.Flags().BoolVar(&sCfg.skipCollection, "skip-collection", false, "download full collection")
	cmd.Flags().BoolVar(&sCfg.cleanOnStart, "cleanOnStart", true, "clean folders")
	cmd.Flags().IntVar(&sCfg.imgWorker, "img-worker", 10, "nof img proccesing worker")
	cmd.Flags().IntVar(&sCfg.vidWorker, "vid-worker", 5, "nof vid proccesing worker")
	cmd.Flags().IntVar(&sCfg.redWorker, "reddit-worker", 15, "nof reddit proccesing worker")

	return cmd
}

func (cfg *ScrapeCfg) sanitize() error {
	if cfg.dstDir == "" {
		cfg.dstDir = "./download"
	}
	if cfg.scrapeOpts == nil {
		cfg.scrapeOpts = &sources.ScrapeOpts{
			Limit: 25,
			Duration: "day",
		}
	}
	return nil
}

func NewScrapper(cfg *ScrapeCfg) (scr *Scrapper, err error) {
	err = cfg.sanitize()
	if err != nil {
		return nil, err
	}
	scr = &Scrapper{
		sCfg:       &sCfg,
		ctx:        context.Background(),
		dstPath: &DstPath{
			BasePath:   sCfg.dstDir,
			ImgPath:    sCfg.imgDir,
			VidPath:    sCfg.vidDir,
			CombineDir: sCfg.combineDir,
		},
	}
	if len(sCfg.sourceIds) > 0 {
		log.Info("using flags to scrape %s", sCfg.sourceIds)
		scr.SourceAc = sCfg.sourceIds
	} else {
		ReadFromJson(sCfg.subreddits, &scr.SourceAc)
	}
	scr.SourceStore, err = sources.NewRedditStore(scr.ctx, &sources.RedditStoreOpts{
		RedditClientOpts: reddit.RedditClientOpts{
			CfgPath:        sCfg.authCfg,
			SkipCollection: sCfg.skipCollection,
		},
	})
	if err != nil {
		return nil, err
	}
	//creating dir struct
	scr.DstStore = store.FileStore{Dir: sCfg.dstDir}
	return scr, nil
}

func (s Scrapper) createStructure() {
	if sCfg.cleanOnStart {
		err := s.DstStore.CleanAll(s.dstPath.GetBasePath())
		if err != nil {
			log.Warn("err while deleting dir structure ", "error", err)
		} else {
			log.Info("cleanup success")
		}
	}
	for _, f := range s.SourceAc {
		log.Info("creating ", "path", s.dstPath.GetImgPath(f))
		log.Info("creating ", "path", s.dstPath.GetVidPath(f))
		s.DstStore.CreateDir(s.dstPath.GetImgPath(f))
		s.DstStore.CreateDir(s.dstPath.GetVidPath(f))
	}
}
func (s Scrapper) Start() (err error) {
	scr, err := NewScrapper(&sCfg)
	if err != nil {
		return err
	}
	scr.createStructure()
	scr.Run()
	log.Info("Summary", "Processed Imgs :", imgCounter)
	log.Info("Summary", "Processed vids :", vidCounter)
	return err
}

func (s Scrapper) processImg(j Job) {
	//download file
	data, err := s.SourceStore.DownloadJob(j)
	if err != nil {
		log.Warn("failed while downloading imgs", "error", err)
		return
	}

	//save to dir
	log.Info("saving file to filesystem", "filename", j.Name)
	err = s.DstStore.Write(filepath.Join(j.Dst, j.FileName), data)
	if err != nil {
		log.Error("err", fmt.Sprint("failed to save file %s to %s as %s", j.FileName, j.Dst, err))
		return
	}
	atomic.AddInt64(&imgCounter, 1)
}

func (s Scrapper) processVid(j Job) {
	data, err := s.SourceStore.DownloadJob(j)
	if err != nil {
		log.Warn("failed while downloading imgs", "error", err)
		return
	}

	//save to dir
	log.Info("saving file to filesystem", "filename", j.Name)
	err = s.DstStore.Write(filepath.Join(j.Dst, j.FileName), data)
	if err != nil {
		log.Error("err", fmt.Sprint("failed to save file %s to %s as %s", j.FileName, j.Dst, err))
		return
	}
	atomic.AddInt64(&vidCounter, 1)
}

func (s Scrapper) subWorker(id int, m *Mediums, wg *sync.WaitGroup) {
	defer wg.Done()
	wg.Add(1)
	log.Info("started sub worker", "id", id)
	for r := range m.subq {
		p, err := s.SourceStore.ScrapePosts(r, *s.sCfg.scrapeOpts)
		if err != nil {
			log.Error("Error while scraping", "source", r)
			continue
		}
		for post := range p {
			m.postq <- post
		}
	}
	fmt.Println("sub worker exits ", id)
}

func (s Scrapper) imgWorker(id int, m *Mediums) {
	defer m.swg.Done()
	fmt.Println("starting img woker ", id)
	for j := range m.imgq {
		fmt.Println("processing img ", j.Name)
		s.processImg(j)
	}
	fmt.Println("Exited img worker ", id)
}

func (s Scrapper) vidWorker(id int, m *Mediums) {
	defer m.swg.Done()
	fmt.Println("starting vid woker ", id)
	for j := range m.vidq {
		fmt.Println("processing VIDEO ", j.Name, j.Src)
		s.processVid(j)
	}
	fmt.Println("Exited vid worker ", id)

}

func (s Scrapper) startWorkers(m *Mediums) {
	var sub_wg sync.WaitGroup

	for i := range sCfg.redWorker {
		go s.subWorker(i, m, &sub_wg)
	}

	for i := range sCfg.imgWorker {
		m.swg.Add(1)
		go s.imgWorker(i, m)
	}

	for i := range sCfg.vidWorker {
		m.swg.Add(1)
		go s.vidWorker(i, m)
	}

	sub_wg.Wait()
	close(m.postq)
}

func (s Scrapper) Run() {
	var mwg sync.WaitGroup
	m := &Mediums{
		subq:  make(chan string),
		postq: make(chan Post),
		imgq:  make(chan Job, 10),
		vidq:  make(chan Job, 1000),
	}
	go s.startWorkers(m)
	mwg.Add(1)
	go func(wg *sync.WaitGroup) {
		defer func() {
			close(m.subq)
		}()
		for _, sub := range s.SourceAc {
			fmt.Println("scrapping ", sub)
			m.subq <- sub
		}
	}(&mwg)
LOOP:
	for {
		select {
		case v, ok := <-m.postq:
			if !ok {
				close(m.imgq)
				close(m.vidq)
				break LOOP
			}
			if v.MediaType == VID_TYPE {
				if !sCfg.skipVideo {
					m.vidq <- Job{
						Id:        v.Id,
						Src:       v.SrcLink,
						Dst:       s.dstPath.GetVidPath(v.SourceAc),
						Name:      fmt.Sprintf("%s.%s", v.Title, v.Ext),
						MediaType: v.MediaType,
						FileName:  fmt.Sprintf("%s.%s", v.Id, v.Ext),
					}
				}
			}

			if v.MediaType == IMG_TYPE {
				m.imgq <- Job{
					Id:        v.Id,
					Src:       v.SrcLink,
					Dst:       s.dstPath.GetImgPath(v.SourceAc),
					Name:      fmt.Sprintf("%s.%s", v.Title, v.Ext),
					MediaType: v.MediaType,
					FileName:  fmt.Sprintf("%s.%s", v.Id, v.Ext),
				}
			}
		}
	}
	m.swg.Wait()
}
