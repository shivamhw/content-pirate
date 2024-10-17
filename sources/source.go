package sources

import (
	"github.com/shivamhw/reddit-pirate/commons"
)

type Source interface {
	Emit([]string) <-chan commons.Post
}
