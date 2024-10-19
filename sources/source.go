package sources

import (
	"github.com/shivamhw/reddit-pirate/commons"
)

type Source interface {
	Scrape(string, chan<- commons.Post) 
}
