package scrapper

import (
	"context"
	"encoding/json"
	"fmt"
	log "log/slog"
	"sync"
	"sync/atomic"
	"time"

	"github.com/google/uuid"
	"github.com/shivamhw/content-pirate/commons"
	"github.com/shivamhw/content-pirate/pkg/kv"
	"github.com/shivamhw/content-pirate/pkg/reddit"
	"github.com/shivamhw/content-pirate/sources"
	"github.com/shivamhw/content-pirate/store"
)

type ScrapperV1 struct {
	SourceStore sources.Source
	sCfg        *ScrapeCfg
	ctx         context.Context
	M           *Mediums
	swg         sync.WaitGroup
	kv          kv.KV
	l           *sync.Mutex
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
}

type Mediums struct {
	TaskQ chan *Task
	ItemQ chan sources.Item
	imgq  chan sources.Item
	vidq  chan sources.Item
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
		ItemQ: make(chan sources.Item),
		imgq:  make(chan sources.Item, 10),
		vidq:  make(chan sources.Item, 10),
	}
	scr = &ScrapperV1{
		sCfg: cfg,
		ctx:  context.Background(),
		M:    m,
		kv:   kv.GetInMemoryKv(),
		l:    &sync.Mutex{},
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

func (s *ScrapperV1) getStore(d *store.DstPath) (store.Store, error) {
	if store, err := store.NewFileStore(d); err != nil {
		return nil, err
	} else {
		return store, nil
	}
}

func (s *ScrapperV1) processImg(i sources.Item) {
	//download file
	defer s.increment(i.TaskId)
	data, err := s.SourceStore.DownloadItem(i)
	if err != nil {
		log.Warn("failed while downloading imgs", "error", err)
		return
	}

	//save to dir
	log.Info("saving file to filesystem", "dst", i.Dst)

	i.Dst, err = i.DstStore.Write(i.Dst, commons.IMG_TYPE, data)
	if err != nil {
		log.Error("err", fmt.Sprint("failed to save file %s to %s as %s", i.FileName, i.Dst, err))
		return
	}

	atomic.AddInt64(&imgCounter, 1)
}

func (s *ScrapperV1) increment(id string) {

	log.Debug("incrementing item done", "taskId", id)
	t, err := s.GetJob(id)
	if err != nil {
		log.Error("error incrementing", "taskId", id)
		return
	}
	atomic.AddInt64(&t.Status.ItemDone, 1)
	_, err = s.UpdateItemDone(id, TaskUpdateOpts{
		TaskStatus: &t.Status,
	})
	if err != nil {
		log.Error("error incrementing", "taskId", id)
		return
	}
}

func (s *ScrapperV1) processVid(i sources.Item) {
	data, err := s.SourceStore.DownloadItem(i)
	if err != nil {
		log.Warn("failed while downloading imgs", "error", err)
		return
	}

	//save to dir
	log.Info("saving file to filesystem", "dst", i.Dst)
	i.Dst, err = i.DstStore.Write(i.Dst, commons.VID_TYPE, data)
	if err != nil {
		log.Error("err", fmt.Sprint("failed to save file %s to %s as %s", i.FileName, i.Dst, err))
		return
	}
	atomic.AddInt64(&vidCounter, 1)
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
			log.Info("Scrapping", "src", v)
			p, err := s.SourceStore.ScrapePosts(v.J.SrcAc, sources.ScrapeOpts(v.J.Opts))
			if err != nil {
				log.Error("Error while scraping", "source", v)
				continue
			}
			wg.Add(1)
			go func(wg *sync.WaitGroup) {
				defer wg.Done()
				for post := range p {
					fileName := fmt.Sprintf("%s.%s", post.Id, post.Ext)
					dst := fmt.Sprintf("%s.%s", post.Id, post.Ext)
					if !v.J.Dst.CombineDir {
						dst = fmt.Sprintf("%s/%s", v.J.SrcAc, dst)
					}

					item := sources.Item{
						Id:       post.Id,
						TaskId:   v.Id,
						Src:      post.SrcLink,
						Title:    post.Title,
						FileName: fileName, 
						Dst:      dst,
						Type:     post.MediaType,
						Ext:      post.Ext,
						SourceAc: post.SourceAc,
						DstStore: v.S,
					}
					v.I = append(v.I, item)
					if item.DstStore.FileExists(item.Dst, post.MediaType){
						log.Warn("file exists not adding it to queue", "file", item.Dst)
						s.increment(item.TaskId)
						continue
					}
					s.M.ItemQ <- item
				}
				//todo: fix update of total element
				v.Status.TotalItem = int64(len(v.I))
				log.Info("updating total item", "task", v.Id, "items", v.Status.TotalItem)
				v.Status.Status = TaskStarted
				nTask, err := s.UpdateTask(v.Id, TaskUpdateOpts{
					TaskStatus: &v.Status,
					Items: v.I,
				})
				v.Status = nTask.Status
				if err != nil {
					log.Error("updating status of task failed", "id", v.Id)
				}
			}(&wg)
		case <-t.C:
			{
				log.Info("waiting for topics to scrape")
			}
		}
	}
	log.Warn("topic closed, waiting for routines to feed posts")
	wg.Wait()
	s.M.closeAll()
	log.Warn("stopped recieving topics to scrape... exiting")
}

func (s *ScrapperV1) imgWorker(id int) {
	defer s.swg.Done()
	fmt.Println("starting img woker ", id)
	for j := range s.M.imgq {
		log.Debug("processing img ","title", j.Title)
		s.processImg(j)
	}
	fmt.Println("Exited img worker ", id)
}

func (s *ScrapperV1) vidWorker(id int) {
	defer s.swg.Done()
	fmt.Println("starting vid woker ", id)
	for j := range s.M.vidq {
		log.Debug("processing VID ","title", j.Title)
		s.processVid(j)
	}
	fmt.Println("Exited vid worker ", id)
}

func (s *ScrapperV1) startWorkers() {
	for range s.sCfg.TopicWorkers {
		go s.subWorker()
	}

	for i := range s.sCfg.ImgWorkers {
		s.swg.Add(1)
		go s.imgWorker(i)
	}

	for i := range s.sCfg.VidWorkers {
		s.swg.Add(1)
		go s.vidWorker(i)
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
			if v.Type == commons.VID_TYPE {
				s.M.vidq <- v
			}

			if v.Type == commons.IMG_TYPE {
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

func (s *ScrapperV1) SubmitJob(j Job) (id string, err error) {
	id = uuid.NewString()
	//create task from job
	store, err := s.getStore(&j.Dst)
	if err != nil {
		return "", err
	}
	t := Task{
		Id: id,
		J:  j,
		I:  []sources.Item{},
		Status: TaskStatus{
			ItemDone:  0,
			TotalItem: 0,
			Status:    TaskCreated,
		},
		S: store,
	}
	//put task to queue
	log.Info("submitting task ", "task", t)
	data, _ := json.Marshal(t)
	err = s.kv.Set("task", id, data)
	if err != nil {
		return "", err
	}
	s.M.TaskQ <- &t
	return id, nil
}

func (s *ScrapperV1) GetJob(id string) (Task, error) {
	var t Task
	data, err := s.kv.Get("task", id)
	if err != nil {
		return Task{}, err
	}
	err = json.Unmarshal(data, &t)
	if err != nil {
		log.Error(err.Error())
		return Task{}, err
	}
	return t, nil
}

func (s *ScrapperV1) CheckJob(id string) (TaskStatus, error) {
	t, err := s.GetJob(id)
	if err != nil {
		return TaskStatus{}, err
	}
	return t.Status, nil
}

func (s *ScrapperV1) WaitOnId(id string) {
	//check if id is done
	log.Info("waiting to complete", "id", id)
	for {
		s, err := s.CheckJob(id)
		if err != nil {
			return
		}
		log.Info("status", "task", id, "Completed", s.ItemDone, "Total", s.TotalItem)
		time.Sleep(5 * time.Second)
		if s.ItemDone >= s.TotalItem && s.Status != TaskCreated {
			break
		}
	}
}

func (s *ScrapperV1) UpdateTask(id string, opts TaskUpdateOpts) (Task, error) {
	defer s.l.Unlock()
	s.l.Lock()
	var t Task
	data, err := s.kv.Get("task", id)
	if err != nil {
		return Task{}, err
	}
	err = json.Unmarshal(data, &t)
	if err != nil {
		return Task{}, err
	}
	if opts.TaskStatus != nil {
		t.Status.TotalItem = opts.TaskStatus.TotalItem
		t.Status.Status = opts.TaskStatus.Status
	}
	if opts.Items != nil {
		t.I = append(t.I, opts.Items...)
	}
	// hack alert
	v, _ := json.Marshal(t)
	err = s.kv.Set("task", id, v)
	if err != nil {
		return Task{}, err
	}
	return t, nil
}

func (s *ScrapperV1) UpdateItemDone(id string, opts TaskUpdateOpts) (Task, error) {
	defer s.l.Unlock()
	s.l.Lock()
	var t Task
	data, err := s.kv.Get("task", id)
	if err != nil {
		return Task{}, err
	}
	err = json.Unmarshal(data, &t)
	if err != nil {
		return Task{}, err
	}
	t.Status.ItemDone = opts.ItemDone
	// hack alert
	v, _ := json.Marshal(t)
	err = s.kv.Set("task", id, v)
	if err != nil {
		return Task{}, err
	}
	return t, nil
}
