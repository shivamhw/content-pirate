package telegram

import (
	"context"
	"fmt"
	log "log/slog"
	"math/rand"
	"time"

	"github.com/gotd/contrib/bg"
	"github.com/gotd/td/telegram"
	"github.com/gotd/td/telegram/auth"
	"github.com/gotd/td/tg"
	"github.com/iyear/tdl/core/logctx"
	"github.com/iyear/tdl/pkg/tclient"
)

type Telegram struct {
	ctx   context.Context
	user  *UserData
	c     *telegram.Client
	close *bg.StopFunc
}

type UserData struct {
	PhoneNumber string
	Store       *Store
}

type Recipient struct {
	AccessHash int64
	UserId     int64
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
	stop, err := bg.Connect(client)
	if err != nil {
		return nil, err
	}
	return &Telegram{
		ctx:   ctx,
		user:  user,
		c:     client,
		close: &stop,
	}, nil
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
	return c, nil
}

func (t *Telegram) ListChats() (result []*Dialog, err error) {
	result, err = List(logctx.Named(t.ctx, "ls"), t.c, t.user.Store.Kvd, ListOptions{Filter: "true"})
	if err != nil {
		return result, err
	}
	for _, r := range result {
		log.Info(r.VisibleName)
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


func (t *Telegram) GetChatHistory(chat *Recipient, opts *SearchOpts) (result []*tg.Message, err error) {
	peer := &tg.InputPeerUser{
		UserID:     chat.UserId,
		AccessHash: chat.AccessHash,
	}
	opts.Peer = peer
	his, err := t.c.API().MessagesGetHistory(t.ctx, opts)
	if err != nil {
		return result, err
	}
	for _, m := range his.(*tg.MessagesMessagesSlice).Messages {
		t, ok := m.(*tg.Message)
		if !ok {
			log.Error("cant convert to msg", "hist", "m")
			return result, nil
		}
		result = append(result, t)
	}
	return result, nil
}

func (t *Telegram) ClickBtn(chat *Recipient, msgId int, btnId []byte) (resp *tg.MessagesBotCallbackAnswer, err error) {
	peer := &tg.InputPeerUser{
		UserID:     chat.UserId,
		AccessHash: chat.AccessHash,
	}
	resp, err = t.c.API().MessagesGetBotCallbackAnswer(t.ctx, &tg.MessagesGetBotCallbackAnswerRequest{
		Peer:  peer,
		MsgID: msgId,
		Data:  btnId,
	})
	if err != nil {
		log.Error("click failed: %s", err.Error())
		return nil, err
	}

	log.Info("Callback response:", resp)
	return resp, nil
}

func (t *Telegram) SendMsg(to *Recipient, msg string) (nMsg *tg.Message, err error) {
	peer := &tg.InputPeerUser{
		UserID:     to.UserId,
		AccessHash: to.AccessHash,
	}
	res, err := t.c.API().MessagesSendMessage(t.ctx, &tg.MessagesSendMessageRequest{
		Peer:     peer,
		Message:  msg,
		RandomID: rand.Int63(),
	})
	if err != nil {
		log.Error("err", "e", err)
		return nil, err
	}
	nMsg = extractSentMessage(res)
	return nMsg, err
}


func (t *Telegram) ForwardMsg(from *Recipient, to *Recipient, msgId int) (nMsg *tg.Message, err error) {
	resp, err := t.c.API().MessagesForwardMessages(t.ctx, &tg.MessagesForwardMessagesRequest{
		FromPeer: &tg.InputPeerUser{
			UserID:     from.UserId,
			AccessHash: from.AccessHash,
		},
		ToPeer: &tg.InputPeerChannel{
			ChannelID:  to.UserId,
			AccessHash: to.AccessHash,
		},
		ID:         []int{msgId},
		RandomID:   []int64{rand.Int63()},
		DropAuthor: true, // ✅ Hides "Forwarded from ..."
	})
	if err != nil {
		return nil, err
	}
	nMsg = extractSentMessage(resp)
	return nMsg, nil
}
