package telegram

import (
	"context"
	"fmt"
	"math/rand"
	"strconv"
	"time"

	"github.com/shivamhw/content-pirate/pkg/log"
	"github.com/gotd/contrib/bg"
	"github.com/gotd/td/telegram"
	"github.com/gotd/td/telegram/auth"
	"github.com/gotd/td/telegram/peers"
	"github.com/gotd/td/tg"
	"github.com/iyear/tdl/core/logctx"
	"github.com/iyear/tdl/core/storage"
	"github.com/iyear/tdl/core/tmedia"
	"github.com/iyear/tdl/core/util/tutil"
	"github.com/iyear/tdl/pkg/tclient"
)

type Telegram struct {
	ctx     context.Context
	user    *UserData
	c       *telegram.Client
	close   bg.StopFunc
	manager *peers.Manager
	store   *Store
}

type UserData struct {
	PhoneNumber string
	Store       *Store
}

type Recipient struct {
	UserId int64
}

type SearchOpts = tg.MessagesGetHistoryRequest

func NewTelegram(ctx context.Context, user *UserData) (*Telegram, error) {
	store, err := GetOrCreateStore(ctx, user.PhoneNumber)
	if err != nil {
		return nil, err
	}
	user.Store = store
	client, err := GetClientWithStore(ctx, store)
	if err != nil {
		return nil, err
	}

	manager := peers.Options{Storage: storage.NewPeers(store.Kvd)}.Build(client.API())
	t := &Telegram{
		ctx:  ctx,
		user: user,
		c:    client,
		store: store,
		manager: manager,
	}

	go t.heartBeat()

	return t, nil
}

func (t *Telegram) heartBeat() error {
	ti := time.NewTicker(5 * time.Second)
	for {
		select {
		case <-ti.C:
			log.Infof("telegram heartbeat")
			if err := t.c.Ping(t.ctx); err != nil {
				log.Warnf("telegram reconncting")
				c, err := GetClientWithStore(t.ctx, t.store)
				if err != nil {
					log.Errorf("reconnect failed with telegram")
					panic(err)
				}
				t.c = c
				log.Infof("reconnect successfull")
			}

		}
	}
}

func (t *Telegram) WhoAmI() (status *auth.Status, err error) {
	status, err = t.c.Auth().Status(t.ctx)
	if err != nil {
		return nil, err
	}
	return status, err
}

func GetClientWithStore(ctx context.Context, store *Store) (*telegram.Client, error) {
	c, err := tclient.New(ctx, tclient.Options{
		KV:               store.Kvd,
		UpdateHandler:    nil,
		ReconnectTimeout: 1 * time.Minute,
	}, false)
	if err != nil {
		return nil, err
	}

	_, err = bg.Connect(c)
	if err != nil {
		return nil, err
	}
	return c, nil
}

func (t *Telegram) ListChats() (result []*Dialog, err error) {
	result, err = List(logctx.Named(t.ctx, "ls"), t.c, t.user.Store.Kvd, ListOptions{Filter: "true"})
	if err != nil {
		return result, err
	}
	for _, r := range result {
		log.Infof(r.VisibleName)
	}
	return result, nil
}

func (t *Telegram) SearchChats(q string) (result []*Dialog, err error) {
	resolved, err := t.c.API().ContactsSearch(t.ctx, &tg.ContactsSearchRequest{
		Q:     q,
		Limit: 5,
	})
	for _, chat := range resolved.Chats {
		switch c := chat.(type) {
		case *tg.Channel:
			r := &Dialog{
				ID:          c.ID,
				AccessHash:  c.AccessHash,
				Type:        DialogChannel,
				Username:    c.Username,
				VisibleName: c.Title,
			}
			result = append(result, r)
		}
	}
	return result, err

}

func (t *Telegram) GetUserFromUsername(username string) (user *tg.User, err error) {
	res, err := t.c.API().ContactsResolveUsername(t.ctx, &tg.ContactsResolveUsernameRequest{
		Username: username,
	})
	if err != nil {
		return nil, err
	}
	if res == nil {
		return nil, fmt.Errorf("user not found %s", username)
	}
	return res.Users[0].(*tg.User), nil
}

func (t *Telegram) SearchUsers(q string) (result []*Dialog, err error) {
	resolved, err := t.c.API().ContactsSearch(t.ctx, &tg.ContactsSearchRequest{
		Q:     q,
		Limit: 5,
	})
	for _, chat := range resolved.Users {
		switch c := chat.(type) {
		case *tg.User:
			r := &Dialog{
				ID:          c.ID,
				AccessHash:  c.AccessHash,
				Type:        DialogPrivate,
				Username:    c.Username,
				VisibleName: c.FirstName,
			}
			result = append(result, r)
		}
	}
	return result, err

}

func (t *Telegram) GetChatHistory(chat *Recipient, opts *SearchOpts) (result []tg.Message, err error) {
	peer, err := tutil.GetInputPeer(t.ctx, t.manager, fmt.Sprintf("%d", chat.UserId))
	if err != nil {
		return nil, err
	}
	opts.Peer = peer.InputPeer()
	his, err := t.c.API().MessagesGetHistory(t.ctx, opts)
	if err != nil {
		return result, err
	}

	switch v := his.(type) {
	case *tg.MessagesMessages:
		for _, msg := range v.Messages {
			if m, ok := msg.(*tg.Message); ok {
				result = append(result, *m)
			}
		}

	case *tg.MessagesMessagesSlice:
		for _, msg := range v.Messages {
			if m, ok := msg.(*tg.Message); ok {
				result = append(result, *m)
			}
		}

	case *tg.MessagesChannelMessages:
		for _, msg := range v.Messages {
			if m, ok := msg.(*tg.Message); ok {
				result = append(result, *m)
			}
		}

	case *tg.MessagesMessagesNotModified:
		// No new messages, return empty result
		return nil, nil

	default:
		panic(fmt.Sprintf("unexpected response type: %T", v))
	}

	return result, nil
}

func (t *Telegram) ClickBtn(chat *Recipient, msgId int, btnId []byte) (resp *tg.MessagesBotCallbackAnswer, err error) {
	peer, err := tutil.GetInputPeer(t.ctx, t.manager, fmt.Sprintf("%d", chat.UserId))
	if err != nil {
		return nil, err
	}
	resp, err = t.c.API().MessagesGetBotCallbackAnswer(t.ctx, &tg.MessagesGetBotCallbackAnswerRequest{
		Peer:  peer.InputPeer(),
		MsgID: msgId,
		Data:  btnId,
	})
	if err != nil {
		log.Errorf("click failed: %s", err.Error())
		return nil, err
	}

	log.Debugf("Callback response:", resp)
	return resp, nil
}

func (t *Telegram) SendMsg(to *Recipient, msg string) (nMsg *tg.Message, err error) {
	peer, err := tutil.GetInputPeer(t.ctx, t.manager, fmt.Sprintf("%d", to.UserId))
	if err != nil {
		return nil, err
	}
	res, err := t.c.API().MessagesSendMessage(t.ctx, &tg.MessagesSendMessageRequest{
		Peer:     peer.InputPeer(),
		Message:  msg,
		RandomID: rand.Int63(),
	})
	if err != nil {
		log.Errorf("err", "e", err)
		return nil, err
	}
	nMsg = extractSentMessage(res)
	return nMsg, err
}

func (t *Telegram) ForwardMsg(from string, to string, msg string) (nMsg *tg.Message, err error) {
	fromPeer, err := tutil.GetInputPeer(t.ctx, t.manager, from)
	if err != nil {
		return nil, err
	}
	toPeer, err := tutil.GetInputPeer(t.ctx, t.manager, to)
	if err != nil {
		return nil, err
	}
	msgId, err := strconv.Atoi(msg)
	if err != nil {
		return nil, err
	}
	resp, err := t.c.API().MessagesForwardMessages(t.ctx, &tg.MessagesForwardMessagesRequest{
		FromPeer:   fromPeer.InputPeer(),
		ToPeer:     toPeer.InputPeer(),
		ID:         []int{msgId},
		RandomID:   []int64{rand.Int63()},
		DropAuthor: true,
	})
	if err != nil {
		return nil, err
	}
	nMsg = extractSentMessage(resp)
	return nMsg, nil
}

func GetFilenameFromMessage(msg *tg.Message) string {
	media, ok := msg.GetMedia()
	id := fmt.Sprintf("%d", msg.ID)
	if !ok {
		return id
	}
	mm, ok := tmedia.ExtractMedia(media)
	if !ok {
		return id
	}
	return mm.Name
}
