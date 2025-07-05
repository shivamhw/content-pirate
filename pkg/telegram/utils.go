package telegram

import (
	"github.com/gotd/td/tg"
)

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