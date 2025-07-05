package scrapper

import (
	"context"
	"fmt"
	log "log/slog"
	"sync"
	"sync/atomic"
	"time"

	"github.com/shivamhw/content-pirate/commons"
	"github.com/shivamhw/content-pirate/pkg/kv"
	"github.com/shivamhw/content-pirate/pkg/reddit"
	"github.com/shivamhw/content-pirate/sources"
	"github.com/shivamhw/content-pirate/store"
)

type ScrapperV1 struct {
	SourceStore  sources.Source
	sCfg         *ScrapeCfg
	ctx          context.Context
	M            *Mediums
	swg          sync.WaitGroup
	kv           kv.KV
	l            *sync.Mutex
	taskStoreIdx map[string][]store.Store
}

type AuthCfg struct {
	ID       string `json:"id"`
	Secret   string `json:"secret"`
	Username string `json:"username"`
	Password string `json:"password"`
}

type ScrapeCfg struct {
	AuthCfg      string
	ImgWorkers   int
	VidWorkers   int
	TopicWorkers int
	TimeOut      int64 //in seconds
	SourceType   sources.SourceType
}

type Mediums struct {
	TaskQ chan *Task
	ItemQ chan DownloadItemJob
	imgq  chan DownloadItemJob
	vidq  chan DownloadItemJob
}

var (
	imgCounter int64
	vidCounter int64
)

func NewScrapper(cfg *ScrapeCfg) (scr *ScrapperV1, err error) {
	err = cfg.sanitize()
	if err != nil {
		return nil, err
	}

	//creating mediums
	m := &Mediums{
		TaskQ: make(chan *Task),
		ItemQ: make(chan DownloadItemJob),
		imgq:  make(chan DownloadItemJob, 10),
		vidq:  make(chan DownloadItemJob, 10),
	}
	scr = &ScrapperV1{
		sCfg:         cfg,
		ctx:          context.Background(),
		M:            m,
		kv:           kv.GetInMemoryKv(),
		l:            &sync.Mutex{},
		taskStoreIdx: make(map[string][]store.Store),
	}

	scr.SourceStore, err = sources.NewRedditStore(scr.ctx, &sources.RedditStoreOpts{
		RedditClientOpts: reddit.RedditClientOpts{
			CfgPath: cfg.AuthCfg,
		},
	})
	if err != nil {
		return nil, err
	}
	return scr, nil
}

func (s *ScrapperV1) process(i *DownloadItemJob) {
	//download file
	defer s.increment(i.T.Id)
	err := s.SourceStore.DownloadItem(i.I.Ctx, i.I)
	if err != nil {
		log.Warn("failed while downloading", "name", i.I.FileName, "error", err)
		return
	}

	//save to dir
	log.Debug("saving file to filesystem", "dst", i.I.Dst)

	if err:= s.saveItem(i); err != nil {
		log.Error("error saving", "item", i.I.FileName, "err", err)
	}
	atomic.AddInt64(&imgCounter, 1)
}

func (s *ScrapperV1) saveItem(i *DownloadItemJob) (err error) {
	for _, s := range i.stores {
		if dst, err := s.Write(i.I); err != nil {
			return err
		} else {

			i.I.Dst = dst
		}
	}
	return
}

func (s *ScrapperV1) subWorker() {
	t := time.NewTicker(5 * time.Second)
	wg := sync.WaitGroup{}
LOOP:
	for {
		select {
		case v, ok := <-s.M.TaskQ:
			if !ok {
				break LOOP
			}
			log.Debug("Scrapping", "src", v)
			p, err := s.SourceStore.ScrapePosts(s.ctx, v.J.SrcAc, sources.ScrapeOpts(v.J.Opts))
			if err != nil {
				log.Error("Error while scraping", "source", v)
				continue
			}
			wg.Add(1)
			go func(wg *sync.WaitGroup) {
				defer wg.Done()
				for post := range p {
					fileName := fmt.Sprintf("%s.%s", post.Id, post.Ext)
					dst := fmt.Sprintf("%s/%s.%s",v.J.SrcAc, post.Id, post.Ext)
					ctx := s.ctx

					if s.sCfg.TimeOut > 0 {
						ctx, _ = context.WithTimeout(ctx, time.Duration(s.sCfg.TimeOut)*time.Second)
					}
					item := commons.Item{
						Id:       post.Id,
						Src:      post.SrcLink,
						Title:    post.Title,
						FileName: fileName,
						Dst:      dst,
						Type:     post.MediaType,
						Ext:      post.Ext,
						SourceAc: post.SourceAc,
						Ctx:      ctx,
					}
					v.I = append(v.I, item)
					v.Status.TotalItem = int64(len(v.I))
					log.Debug("updating total item", "task", v.Id, "items", v.Status.TotalItem)
					v.Status.Status = TaskStarted
					nTask, err := s.UpdateTask(v.Id, TaskUpdateOpts{
						TaskStatus: &v.Status,
						Items:      []commons.Item{item},
					})
					v.Status = nTask.Status
					if err != nil {
						log.Error("updating status of task failed", "id", v.Id)
					}
					stores := s.filterStores(v, &item)
					if len(stores) <= 0 {
						log.Warn("file exists in all stores not adding it to queue", "file", item.Dst)
						s.increment(v.Id)
						continue
					}
					s.M.ItemQ <- DownloadItemJob{
						I: &item,
						T: v,
						stores: stores,
					}
				}
			}(&wg)
		case <-t.C:
			{
				log.Info("scrapper heartbeat......")
			}
		}
	}
	log.Warn("topic closed, waiting for routines to feed posts")
	wg.Wait()
	s.M.closeAll()
	log.Warn("stopped recieving topics to scrape... exiting")
}

func (s *ScrapperV1) filterStores(t *Task, i *commons.Item) (fStores []store.Store) {
	for _, st := range s.taskStoreIdx[t.Id] {
		if st.ItemExists(i) {
			log.Warn("file already exist", "file", i.FileName)
			continue
		}
		fStores = append(fStores, st)
	}
	return 
}

func (s *ScrapperV1) queueWorker(id int, q chan DownloadItemJob) {
	defer s.swg.Done()
	fmt.Println("starting img woker ", id)
	for j := range q {
		log.Debug("processing", "title", j.I.Title)
		s.process(&j)
	}
	fmt.Println("Exited worker ", id)
}

func (s *ScrapperV1) startWorkers() {
	for range s.sCfg.TopicWorkers {
		go s.subWorker()
	}

	for i := range s.sCfg.ImgWorkers {
		s.swg.Add(1)
		go s.queueWorker(i, s.M.imgq)
	}

	for i := range s.sCfg.VidWorkers {
		s.swg.Add(1)
		go s.queueWorker(i, s.M.vidq)
	}
}

func (s *ScrapperV1) Start() {
	//reset counters
	imgCounter, vidCounter = 0, 0

	go s.startWorkers()
LOOP:
	for {
		select {
		case v, ok := <-s.M.ItemQ:
			if !ok {
				close(s.M.imgq)
				close(s.M.vidq)
				break LOOP
			}
			if v.I.Type == commons.VID_TYPE {
				s.M.vidq <- v
			}

			if v.I.Type == commons.IMG_TYPE {
				s.M.imgq <- v
			}
		}
	}
	s.swg.Wait()
	log.Info("Summary", "Processed Imgs :", imgCounter)
	log.Info("Summary", "Processed vids :", vidCounter)
}

func (cfg *ScrapeCfg) sanitize() error {

	if cfg.ImgWorkers <= 0 {
		cfg.ImgWorkers = 5
	}
	if cfg.TopicWorkers <= 0 {
		cfg.TopicWorkers = 5
	}
	if cfg.VidWorkers <= 0 {
		cfg.VidWorkers = 5
	}
	return nil
}

func (m *Mediums) closeAll() {
	close(m.ItemQ)
}

func (s *ScrapperV1) Stop() {
	log.Warn("Stopping scrapper")
}