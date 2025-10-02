package telegram

import (
	"context"
	"fmt"
	"math/rand"
	"os"
	"path/filepath"
	"strconv"
	"strings"
	"time"

	"github.com/celestix/gotgproto"
	"github.com/celestix/gotgproto/sessionMaker"
	"github.com/go-faster/errors"

	"github.com/gotd/td/telegram"
	"github.com/gotd/td/telegram/auth"
	"github.com/gotd/td/telegram/peers"
	"github.com/gotd/td/tg"

	"github.com/iyear/tdl/core/tclient"
	"github.com/iyear/tdl/core/tmedia"
	"github.com/iyear/tdl/core/util/tutil"
	"github.com/shivamhw/content-pirate/pkg/log"
)

var Appid = 15055931
var AppHash = "021d433426cbb920eeb95164498fe3d3"
var SessionDir = "./teleLogin"
var NotAuthorizedErr = errors.New("unauthorized")

type Telegram struct {
	ctx  context.Context
	opts *ClientOpts
	c    *gotgproto.Client
}

type ClientOpts struct {
	Phone            string
	SessionPath      string
	ReconnectTimeout time.Duration
}

type SearchOpts = tg.MessagesGetHistoryRequest

func (c *ClientOpts) Sanitize() error {
	if c.Phone == "" {
		return fmt.Errorf("empty phone number")
	}
	if c.SessionPath == "" {
		abPath, err := filepath.Abs(SessionDir)
		if err != nil {
			return fmt.Errorf("error getting abs path %w", err)
		}
		t := filepath.Join(abPath, c.Phone, "session.json")
		log.Infof("using default session path %s", t)
		c.SessionPath = t
	}
	if c.ReconnectTimeout == 0 {
		c.ReconnectTimeout = time.Second * 5
	}
	if ok, _ := os.Stat(c.SessionPath); ok == nil {
		log.Infof("creating session file")
		_ = os.MkdirAll(filepath.Dir(c.SessionPath), 0755)
		f, err := os.Create(c.SessionPath)
		if err != nil {
			return err
		}
		f.Close()
	}
	return nil
}

func NewTelegram(ctx context.Context, opts *ClientOpts) (*Telegram, error) {
	client, err := gotgproto.NewClient(Appid, AppHash, gotgproto.ClientTypePhone(opts.Phone), &gotgproto.ClientOpts{
		Session:          sessionMaker.JsonFileSession(opts.SessionPath),
		NoAutoAuth:       true,
		DisableCopyright: true,
		Middlewares:      tclient.NewDefaultMiddlewares(ctx, opts.ReconnectTimeout),
	})
	if err != nil {
		if strings.Contains(err.Error(), "session is unauthorized") {
			return nil, NotAuthorizedErr
		}
		return nil, err
	}
	t := &Telegram{
		ctx:  ctx,
		opts: opts,
		c:    client,
	}

	return t, nil
}

func (t *Telegram) WhoAmI() (status *auth.Status, err error) {
	status, err = t.c.Auth().Status(t.ctx)
	if err != nil {
		return nil, fmt.Errorf("error in WhoAmI, err %w", err)
	}
	return status, nil
}

func (t *Telegram) SearchChat(c string, q string) (result []tg.Message, err error) {
	peer, err := tutil.GetInputPeer(t.ctx, peers.Options{}.Build(t.c.API()), c)
	if err != nil {
		return
	}
	res, err := t.c.API().MessagesSearch(t.ctx, &tg.MessagesSearchRequest{
		Q:      q,
		Filter: &tg.InputMessagesFilterEmpty{},
		Peer:   peer.InputPeer(),
	})
	if err != nil {
		return
	}
	result, err = convertMsgcls(res)
	if err != nil {
		return
	}
	return
}

func (t *Telegram) ListChats() (result []*Dialog, err error) {
	return
	// result, err = List(logctx.Named(t.ctx, "ls"), t.c, t.user.Store.Kvd, ListOptions{Filter: "true"})
	// if err != nil {
	// 	return result, err
	// }
	// for _, r := range result {
	// 	log.Infof(r.VisibleName)
	// }
	// return result, nil
}

func (t *Telegram) SearchUsername(q string) (result []*Dialog, err error) {
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
	var res *tg.ContactsResolvedPeer
	err = t.c.Run(context.Background(), func(ctx context.Context) error {
		if res, err = t.c.API().ContactsResolveUsername(t.ctx, &tg.ContactsResolveUsernameRequest{Username: username}); err != nil {
			return err
		}
		return nil
	})
	if res == nil {
		return nil, fmt.Errorf("user not found %s", username)
	}
	return res.Users[0].(*tg.User), err
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

func (t *Telegram) GetChatHistory(chatId string, opts *SearchOpts) (result []tg.Message, err error) {
	peer, err := tutil.GetInputPeer(t.ctx, peers.Options{}.Build(t.c.API()), chatId)
	if err != nil {
		return
	}
	opts.Peer = peer.InputPeer()
	his, err := t.c.API().MessagesGetHistory(t.ctx, opts)
	if err != nil {
		return
	}
	result, err = convertMsgcls(his)
	if err != nil {
		return
	}
	return
}

func convertMsgcls(his tg.MessagesMessagesClass) (result []tg.Message, err error) {

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
	return
}

func (t *Telegram) ClickBtn(chatId string, msgId int, btnId []byte) (resp *tg.MessagesBotCallbackAnswer, err error) {
	peer, err := tutil.GetInputPeer(t.ctx, peers.Options{}.Build(t.c.API()), chatId)
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

func (t *Telegram) SendMsg(to string, msg string) (nMsg *tg.Message, err error) {
	peer, err := tutil.GetInputPeer(t.ctx, peers.Options{}.Build(t.c.API()), to)
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
	fromPeer, err := tutil.GetInputPeer(t.ctx, peers.Options{}.Build(t.c.API()), from)
	if err != nil {
		return nil, err
	}
	toPeer, err := tutil.GetInputPeer(t.ctx, peers.Options{}.Build(t.c.API()), to)
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

func GetMediaFromMessage(msg *tg.Message) (*tmedia.Media, bool) {
	media, ok := msg.GetMedia()
	if !ok {
		return nil, false
	}
	mm, ok := tmedia.ExtractMedia(media)
	if !ok {
		return nil, false
	}
	return mm, true
}

func GetFilenameFromMessage(msg *tg.Message) string {
	mm, ok := GetMediaFromMessage(msg)
	if !ok {
		return fmt.Sprintf("%d", msg.ID)
	}
	return mm.Name
}

func (t *Telegram) GetSingleMessage(msgId int, peer string) (*tg.Message, error) {
	p, err := tutil.GetInputPeer(t.ctx, peers.Options{}.Build(t.c.API()), peer)
	if err != nil {
		return nil, err
	}

	msg, err := tutil.GetSingleMessage(t.ctx, t.c.API(), p.InputPeer(), msgId)
	if err != nil {
		return nil, err
	}

	return msg, nil
}

func (t *Telegram) GetClient() *telegram.Client {
	return t.c.Client
}
