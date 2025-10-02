package sources

import (
	"context"
	"fmt"
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
		Phone: cfg.PhoneNumber,
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
	posts, err := t.scrape(chat, opts)
	if err != nil {
		return nil, err
	}
	go func() {
		defer func() {
			close(post)
		}()
		for _, p := range posts {
			post <- p
		}
	}()
	return
}

func (t *TelegramSource) scrape(src string, opts ScrapeOpts) (p []Post, err error) {
	minId := -1
	lMsgs, err := t.c.GetChatHistory(src, &telegram.SearchOpts{
		Limit: 11,
		OffsetDate: int(opts.LastFrom.Unix()),
	})
	if err != nil {
		log.Errorf(err.Error())
		return nil, err
	}
	if len(lMsgs) <= 0 {
		log.Warnf("no last msg found for", "src", src, "offset", opts.LastFrom.String())
	} else {
		minId = lMsgs[0].ID
	}
	msgs, err := t.c.GetChatHistory(src, &telegram.SearchOpts{
		Limit: opts.Limit,
		MinID: minId,
	})
	log.Infof("scrapped", "unfiltered msgs", len(msgs))
	for _, m := range msgs {
		if m.Date > int(opts.LastFrom.Unix()) {
			log.Debugf("adding msg as the time criteria is met", "msg time", m.Date, "limit", opts.LastFrom.Unix())
			size := 0
			if s, ok := telegram.GetMediaFromMessage(&m); ok {
				size = int(s.Size)
			} else {
				log.Warnf("no media found in message", "msg", m.ID)
			}
			t := Post{
				MediaType: commons.MSG_TYPE,
				Id:        fmt.Sprintf("%d", m.ID),
				SourceAc:  src,
				Title:     m.Message,
				FileName:  telegram.GetFilenameFromMessage(&m),
				Size: int64(size),
			}
			p = append(p, t)
		}
	}
	log.Infof("scrapped", "filtered posts", len(p))
	return
}

func (t *TelegramSource) DownloadItem(ctx context.Context, i *commons.Item) (err error) {
	log.Debugf("downloading", "item", i.Id)
	return
}

func (t *TelegramSource) GetClient() *telegram.Telegram {
	return t.c
}
