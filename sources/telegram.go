package sources

import (
	"context"
	"fmt"
	"strconv"

	"github.com/shivamhw/content-pirate/commons"
	. "github.com/shivamhw/content-pirate/pkg/log"
	"github.com/shivamhw/content-pirate/pkg/telegram"
)

type TelegramSourceOtps struct {
	PhoneNumber string
}

type TelegramSource struct {
	c   *telegram.Telegram
	cfg *TelegramSourceOtps
}

func NewTelegramSource(ctx context.Context, cfg *TelegramSourceOtps) (*TelegramSource, error) {
	user := &telegram.UserData{
		PhoneNumber: cfg.PhoneNumber,
	}
	t, err := telegram.NewTelegram(ctx, user)
	if err != nil {
		return nil, err
	}
	if ok, _ := t.WhoAmI(); !ok.Authorized {
		return nil, fmt.Errorf("user not logged in %s", user.PhoneNumber)
	} else {
		Logger.Info("user logged in ", "user", user.PhoneNumber)
	}
	return &TelegramSource{
		c:   t,
		cfg: cfg,
	}, nil
}

func (t *TelegramSource) ScrapePosts(ctx context.Context, chat string, opts ScrapeOpts) (post chan Post, err error) {
	chatId, err := strconv.ParseInt(chat, 10, 64)
	post = make(chan Post, 5)

	if err != nil {
		return nil, err
	}
	Logger.Info("scrapping telegram ", "id", chatId)
	chatAc := &telegram.Recipient{
		UserId:     chatId,
	}
	posts, err := t.scrape(chatAc, opts)
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

func (t *TelegramSource) scrape(src *telegram.Recipient, opts ScrapeOpts) (p []Post, err error) {
	msgs, err := t.c.GetChatHistory(src, &telegram.SearchOpts{
		Limit: opts.Limit,
	})
	if err != nil {
		Logger.Error(err.Error())
		return nil, err
	}
	Logger.Info("scrapped", "unfiltered msgs", len(msgs))
	for _, m := range msgs {
		if m.Date > int(opts.LastFrom.Unix()) {
			Logger.Debug("adding msg as the time criteria is met", "msg time", m.Date, "limit", opts.LastFrom.Unix())
			t := Post{
				MediaType: commons.MSG_TYPE,
				Id:        fmt.Sprintf("%d", m.ID),
				SourceAc:  fmt.Sprintf("%d", src.UserId),
				Title:     m.Message,
				FileName:  telegram.GetFilenameFromMessage(&m),
			}
			p = append(p, t)
		}
	}
	Logger.Info("scrapped", "filtered posts", len(p))
	return
}

func (t *TelegramSource) DownloadItem(ctx context.Context, i *commons.Item) (err error) {
	Logger.Info("downloading", "item", i.Id)
	return
}

func (t *TelegramSource) GetClient() *telegram.Telegram {
	return t.c
}
