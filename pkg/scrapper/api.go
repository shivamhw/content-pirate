package scrapper

import (
	"encoding/json"
	"sync/atomic"
	"time"

	log "log/slog"

	"github.com/shivamhw/content-pirate/commons"
	"github.com/shivamhw/content-pirate/sources"
	"github.com/shivamhw/content-pirate/store"

	"github.com/google/uuid"
)

type DownloadItemJob struct {
	I      *commons.Item
	T      *Task
	stores []store.Store
}

func (s *ScrapperV1) SubmitJob(j Job) (id string, err error) {
	var stores []store.Store
	id = uuid.NewString()
	//create task from job
	for _, dst := range j.Dst {
		st, err := store.GetStore(dst)
		if err != nil {
			return "", err
		}
		//TODO fix this one on priority
		if s.sCfg.SourceType == sources.SOURCE_TYPE_TELEGRAM {
			log.Warn("using override to add tele client in store")
			st.(*store.TelegramStore).C = s.SourceStore.(*sources.TelegramSource).GetClient()
		}
		stores = append(stores, st)
	}
	t := Task{
		Id: id,
		J:  j,
		I:  []commons.Item{},
		Status: TaskStatus{
			ItemDone:  0,
			TotalItem: 0,
			Status:    TaskCreated,
		},
		S: stores,
	}
	//put task to queue
	log.Info("submitting task ", "task", t)
	data, _ := json.Marshal(t)
	err = s.KV.Set("task", id, data)
	if err != nil {
		return "", err
	}
	s.taskStoreIdx[t.Id] = stores
	s.M.TaskQ <- &t
	return id, nil
}

func (s *ScrapperV1) GetJob(id string) (Task, error) {
	var t Task
	data, err := s.KV.Get("task", id)
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

func (s *ScrapperV1) WaitOnId(id string, waitFor int) bool {
	//check if id is done
	log.Info("waiting to complete", "id", id)
	deadline := time.Now().Add(time.Duration(waitFor) * time.Minute) // time out after 5 mins
	for {
		now := time.Now()
		if now.After(deadline) {
			log.Error("deadline excedded for task", "task", id)
			return false
		}
		s, err := s.CheckJob(id)
		if err != nil {
			return false
		}
		log.Info("status", "task", id, "Completed", s.ItemDone, "Total", s.TotalItem)
		time.Sleep(5 * time.Second)
		if s.ItemDone >= s.TotalItem && s.Status != TaskCreated {
			break
		}
	}
	return true
}

func (s *ScrapperV1) UpdateTask(id string, opts TaskUpdateOpts) (Task, error) {
	defer s.l.Unlock()
	s.l.Lock()
	var t Task
	data, err := s.KV.Get("task", id)
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
	err = s.KV.Set("task", id, v)
	if err != nil {
		return Task{}, err
	}
	return t, nil
}

func (s *ScrapperV1) UpdateItemDone(id string, opts TaskUpdateOpts) (Task, error) {
	defer s.l.Unlock()
	s.l.Lock()
	var t Task
	data, err := s.KV.Get("task", id)
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
	err = s.KV.Set("task", id, v)
	if err != nil {
		return Task{}, err
	}
	return t, nil
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
