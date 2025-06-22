package scrapper

import "github.com/shivamhw/content-pirate/sources"

type Scrapper interface {
	SubmitJob(Job) (string, error)
	CheckJob(string) (TaskStatus, error)
	GetJob(string) (Task, error)
	AddItem(taskId string) (error)
	UpdateItem(itemId string, opts sources.ItemUpdateOpts) (sources.Item, error)
	UpdateTask(taskId string, opts TaskUpdateOpts) (Task, error)
	Stop()
	Start() 
}