package cmd

import (
	"context"
	"fmt"
	log "log/slog"
	"path/filepath"
	"sync"
	"sync/atomic"

	. "github.com/shivamhw/reddit-pirate/commons"
	"github.com/shivamhw/reddit-pirate/pkg/reddit"
	"github.com/shivamhw/reddit-pirate/sources"
	"github.com/shivamhw/reddit-pirate/store"
	"github.com/spf13/cobra"
)

type scrapper struct {
	SourceStore sources.Source
	DstStore    store.Store
	SourceAc    []string
	sCfg        *scrapeCfg
	ctx         context.Context
	dstPath     *DstPath
}

type AuthCfg struct {
	ID       string `json:"id"`
	Secret   string `json:"secret"`
	Username string `json:"username"`
	Password string `json:"password"`
}

type scrapeCfg struct {
	dstDir         string
	vidDir         string
	imgDir         string
	authCfg        string
	subreddits     string
	postId         string
	duration       string
	skipVideo      bool
	cleanOnStart   bool
	combineDir     bool
	skipCollection bool
	imgWorker      int
	vidWorker      int
	redWorker      int
}

type Mediums struct {
	subq  chan string
	postq chan Post
	imgq  chan Job
	vidq  chan Job
	swg   sync.WaitGroup
}

var (
	sCfg       scrapeCfg
	aCfg       AuthCfg
	imgCounter int64
	vidCounter int64
)

func scrapeCmd() *cobra.Command {
	cmd := &cobra.Command{
		Use:   "scrape",
		Long:  "Scrapes subreddit for videos and imgs",
		Short: "scrapes subreddit",
		RunE:   scrapperHandler,
	}
	cmd.Flags().StringVar(&sCfg.dstDir, "dir", "./download", "dst folder for downloads")
	cmd.Flags().StringVar(&sCfg.imgDir, "img-dir", "imgs", "dst folder for imgs")
	cmd.Flags().StringVar(&sCfg.vidDir, "vid-dir", "vids", "dst folder for vids")
	cmd.Flags().StringVar(&sCfg.subreddits, "subs", "./subreddits.json", "list of subreddits")
	cmd.Flags().StringVar(&sCfg.authCfg, "auth", "./reddit.json", "auth config for reddit")
	cmd.Flags().StringVar(&sCfg.postId, "post-id", "", "post id")
	cmd.Flags().StringVar(&sCfg.duration, "duration", "day", "duration")
	cmd.Flags().BoolVar(&sCfg.skipVideo, "skip-vid", true, "skip video download")
	cmd.Flags().BoolVar(&sCfg.combineDir, "combine", true, "combine folders")
	cmd.Flags().BoolVar(&sCfg.skipCollection, "skip-collection", false, "download full collection")
	cmd.Flags().BoolVar(&sCfg.cleanOnStart, "cleanOnStart", true, "clean folders")
	cmd.Flags().IntVar(&sCfg.imgWorker, "img-worker", 10, "nof img proccesing worker")
	cmd.Flags().IntVar(&sCfg.vidWorker, "vid-worker", 5, "nof vid proccesing worker")
	cmd.Flags().IntVar(&sCfg.redWorker, "reddit-worker", 15, "nof reddit proccesing worker")

	return cmd
}

func (s scrapper) createStructure() {
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

func scrapperHandler(cmd *cobra.Command, args []string) (err error){
	scr := scrapper{
		sCfg: &sCfg,
		ctx:  context.Background(),
		dstPath: &DstPath{
			BasePath:   sCfg.dstDir,
			ImgPath:    sCfg.imgDir,
			VidPath:    sCfg.vidDir,
			CombineDir: sCfg.combineDir,
		},
	}
	// load sub reddit
	ReadFromJson(sCfg.subreddits, &scr.SourceAc)
	// create auth
	scr.SourceStore, err = sources.NewRedditStore(scr.ctx, &sources.RedditStoreOpts{
		RedditClientOpts: reddit.RedditClientOpts{
		CfgPath:        sCfg.authCfg,
		SkipCollection: sCfg.skipCollection,
			Duration:       sCfg.duration,
		},
	})
	if err != nil {
		return err
	}
	//creating dir struct
	scr.DstStore = store.FileStore{Dir: sCfg.dstDir}
	scr.createStructure()
	scr.Run()
	log.Info("Summary", "Processed Imgs :", imgCounter)
	log.Info("Summary", "Processed vids :", vidCounter)
	return err
}

func (s scrapper) processImg(j Job) {
	//download file
	data, err := s.SourceStore.DownloadJob(j)
	if err != nil {
		log.Warn("failed while downloading imgs", "error", err)
		return
	}

	//save to dir
	log.Info("saving file to filesystem", "filename", j.Name)
	err = s.DstStore.Write(filepath.Join(j.Dst, j.Name), data)
	if err != nil {
		log.Error("err", fmt.Sprint("failed to save file %s to %s as %s", j.Name, j.Dst, err))
		return
	}
	atomic.AddInt64(&imgCounter, 1)
}

func (s scrapper) processVid(j Job) {
	data, err := s.SourceStore.DownloadJob(j)
	if err != nil {
		log.Warn("failed while downloading imgs", "error", err)
		return
	}

	//save to dir
	log.Info("saving file to filesystem", "filename", j.Name)
	err = s.DstStore.Write(filepath.Join(j.Dst, j.Name), data)
	if err != nil {
		log.Error("err", fmt.Sprint("failed to save file %s to %s as %s", j.Name, j.Dst, err))
		return
	}
	atomic.AddInt64(&vidCounter, 1)
}

func (s scrapper) subWorker(id int, m *Mediums, wg *sync.WaitGroup) {
	defer wg.Done()
	wg.Add(1)
	log.Info("started sub worker", "id", id)
	for r := range m.subq {
		s.SourceStore.ScrapePosts(r, m.postq)
	}
	fmt.Println("sub worker exits ", id)
}

func (s scrapper) imgWorker(id int, m *Mediums) {
	defer m.swg.Done()
	fmt.Println("starting img woker ", id)
	for j := range m.imgq {
		fmt.Println("processing img ", j.Name)
		s.processImg(j)
	}
	fmt.Println("Exited img worker ", id)
}

func (s scrapper) vidWorker(id int, m *Mediums) {
	defer m.swg.Done()
	fmt.Println("starting vid woker ", id)
	for j := range m.vidq {
		fmt.Println("processing VIDEO ", j.Name, j.Src)
		s.processVid(j)
	}
	fmt.Println("Exited vid worker ", id)

}

func (s scrapper) startWorkers(m *Mediums) {
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

func (s scrapper) Run() {
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
						Src:       v.SrcLink,
						Dst:       s.dstPath.GetVidPath(v.SourceAc),
						Name:      fmt.Sprintf("%s_%s.%s", v.Id, v.Title, v.Ext),
						MediaType: v.MediaType,
					}
				}
			}

			if v.MediaType == IMG_TYPE {
				m.imgq <- Job{
					Src:       v.SrcLink,
					Dst:       s.dstPath.GetImgPath(v.SourceAc),
					Name:      fmt.Sprintf("%s_%s.%s", v.Id, v.Title, v.Ext),
					MediaType: v.MediaType,
				}
			}
		}
	}
	m.swg.Wait()
}
