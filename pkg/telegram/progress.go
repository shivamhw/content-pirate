package telegram
import (
"github.com/iyear/tdl/core/downloader"
)
type ProgressTracker struct {
	s string
}

func newProgressTracker() *ProgressTracker {
	return &ProgressTracker{}
}

func (p *ProgressTracker) OnAdd(elem downloader.Elem) {
	return
}

func (p *ProgressTracker) OnDownload(elem downloader.Elem, state downloader.ProgressState) {
	return
}

func (p *ProgressTracker) OnDone(elem downloader.Elem, err error) {
	return
}