package telegram

import (
	"strings"

	"github.com/gotd/td/tg"
	"github.com/iyear/tdl/core/tmedia"
)

const (
	TELEGRAM_VID = 1 << iota
	TELEGRAM_IMG
	TELEGRAM_MSG
	TELEGRAM_DOC
)

type Media struct {
	InputFileLoc tg.InputFileLocationClass
	Name         string
	Size         int64
	DC           int
	Type         int
	Mime         string
}

func extractSentMessage(updates tg.UpdatesClass) *tg.Message {
	up, ok := updates.(*tg.Updates)
	if !ok {
		return nil
	}
	for _, u := range up.Updates {
		switch upd := u.(type) {
		case *tg.UpdateNewMessage:
			if msg, ok := upd.Message.(*tg.Message); ok {
				return msg
			}
		case *tg.UpdateNewChannelMessage:
			if msg, ok := upd.Message.(*tg.Message); ok {
				return msg
			}

		}
	}
	return nil
}

func ParseBtnsFromMsg(msgs *tg.Message) (res map[string][]byte) {
	res = make(map[string][]byte)
	for _, m := range msgs.ReplyMarkup.(*tg.ReplyInlineMarkup).Rows {
		btn := m.Buttons[0]
		txt := btn.(*tg.KeyboardButtonCallback).Text
		data := btn.(*tg.KeyboardButtonCallback).Data
		res[txt] = data
	}
	return
}

func ExtractMedia(m tg.MessageMediaClass) (*Media, bool) {
	switch m := m.(type) {
	case *tg.MessageMediaPhoto:
		m1, ok := tmedia.GetPhotoInfo(m)
		if !ok {
			return nil, ok
		}
		return &Media{
			InputFileLoc: m1.InputFileLoc,
			Name:         m1.Name,
			Size:         m1.Size,
			DC:           m1.DC,
			Type:         TELEGRAM_IMG,
		}, ok
	case *tg.MessageMediaDocument:
		t := TELEGRAM_DOC
		m1, ok := GetDocumentInfo(m)
		if !ok {
			return nil, ok
		}
		switch strings.Split(m1.Mime, "/")[0] {
		case "image":
			t = TELEGRAM_IMG
		case "video":
			t = TELEGRAM_VID
		case "application":
			t = TELEGRAM_DOC
		default:
			t = TELEGRAM_DOC
		}
		return &Media{
			InputFileLoc: m1.InputFileLoc,
			Name:         m1.Name,
			Size:         m1.Size,
			DC:           m1.DC,
			Mime:         m1.Mime,
			Type:         t,
		}, ok
	case *tg.MessageMediaInvoice:
		m1, ok := tmedia.GetExtendedMedia(m.ExtendedMedia)
		if !ok {
			return nil, ok
		}
		return &Media{
			InputFileLoc: m1.InputFileLoc,
			Name:         m1.Name,
			Size:         m1.Size,
			DC:           m1.DC,
			Type:         TELEGRAM_VID,
		}, ok
	}
	return nil, false
}

func GetDocumentInfo(doc *tg.MessageMediaDocument) (*Media, bool) {
	d, ok := doc.Document.(*tg.Document)
	if !ok {
		return nil, false
	}

	return &Media{
		InputFileLoc: &tg.InputDocumentFileLocation{
			ID:            d.ID,
			AccessHash:    d.AccessHash,
			FileReference: d.FileReference,
		},
		Name: tmedia.GetDocumentName(d),
		Size: d.Size,
		DC:   d.DCID,
		Mime: d.MimeType,
	}, true
}
