package sources

import (
	"context"
	"fmt"

	"github.com/gotd/td/tg"
	"github.com/shivamhw/content-pirate/commons"
	"github.com/shivamhw/content-pirate/pkg/log"
	"github.com/shivamhw/content-pirate/pkg/telegram"
)

type TelegramSourceOtps struct {
	PhoneNumber string
	SessionPath string
}

type TelegramSource struct {
	c   *telegram.Telegram
	cfg *TelegramSourceOtps
}

func NewTelegramSource(ctx context.Context, cfg *TelegramSourceOtps) (*TelegramSource, error) {
	opts := &telegram.ClientOpts{
		Phone:       cfg.PhoneNumber,
		SessionPath: cfg.SessionPath,
	}
	if err := opts.Sanitize(); err != nil {
		return nil, err
	}
	t, err := telegram.NewTelegram(ctx, opts)
	if err != nil {
		return nil, err
	}
	if ok, _ := t.WhoAmI(); !ok.Authorized {
		return nil, fmt.Errorf("user not logged in %s", opts.Phone)
	} else {
		log.Infof("user logged in ", "user", opts.Phone)
	}
	return &TelegramSource{
		c:   t,
		cfg: cfg,
	}, nil
}

func (t *TelegramSource) ScrapePosts(ctx context.Context, chat string, opts ScrapeOpts) (post chan Post, err error) {
	post = make(chan Post, 5)

	log.Infof("scrapping telegram ", "id", chat)
	go func() error {
		defer func() {
			close(post)
		}()
	    err := t.scrape(chat, opts, post)
		if err != nil {
			return err
		}
		return nil
	}()
	return
}

func (t *TelegramSource) scrape(src string, opts ScrapeOpts, pChan chan Post) (err error) {
	minId := -1
	if !opts.Full{

	lMsgs, err := t.c.GetChatHistory(src, &telegram.SearchOpts{
		Limit:      11,
		OffsetDate: int(opts.LastFrom.Unix()),
	})
	if err != nil {
		log.Errorf(err.Error())
		return err
	}
	if len(lMsgs) <= 0 {
		log.Warnf("no last msg found for", "src", src, "offset", opts.LastFrom.String())
	} else {
		minId = lMsgs[0].ID
	}
	}
	iter, err := t.c.GetChatHistoryItr(src, &telegram.SearchOpts{
		MinID: minId,
	})
	count := 0
	for iter.Next(context.Background()) {
		log.Debugf("scrapping item ", "c", count)
		msg := iter.Value()
		m, ok := msg.Msg.(*tg.Message)
		if !ok {
			continue
		}
		count++
		t, err := preparePost(src, *m)
		if opts.Limit != 0 && count >= opts.Limit {
			break
		}
		if iter.Err() != nil {	
			return iter.Err()
		}
		log.Debugf("adding msg as the time criteria is met", "msg time", m.Date, "limit", opts.LastFrom.Unix())
		if err != nil {
			log.Errorf("failed converting msg to post", "err", err)
			continue
		}
		pChan <- t
	}
	log.Infof("scrapped", "src" , src, "posts", count)
	return
}

func preparePost(src string, m tg.Message) (p Post, err error) {
	size := 0
	mt := commons.MSG_TYPE
	if m.Media == nil {
		log.Debugf("msg type", mt, "id", m.ID)
	} else {
		s, ok := telegram.GetMediaFromMessage(&m)
		if !ok {
			return p, fmt.Errorf("error parsing media from msg %v",m)
		}
		size = int(s.Size)
		switch s.Type {
		case telegram.TELEGRAM_IMG:
			mt = commons.IMG_TYPE
		case telegram.TELEGRAM_VID:
			mt = commons.VID_TYPE
		case telegram.TELEGRAM_DOC:
			mt = commons.DOC_TYPE
		}
	}
	p = Post{
		MediaType: mt,
		Id:        fmt.Sprintf("%d", m.ID),
		SourceAc:  src,
		Title:     m.Message,
		FileName:  telegram.GetFilenameFromMessage(&m),
		Size:      int64(size),
	}
	return 

}

func (t *TelegramSource) DownloadItem(ctx context.Context, i *commons.Item) (err error) {
	log.Debugf("downloading", "item", i.Id)
	return
}

func (t *TelegramSource) GetClient() *telegram.Telegram {
	return t.c
}
