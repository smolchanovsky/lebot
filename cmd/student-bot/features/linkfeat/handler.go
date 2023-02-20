package linkfeat

import (
	tgbotapi "github.com/go-telegram-bot-api/telegram-bot-api/v5"
	"lebot/cmd/student-bot/core"
	"lebot/cmd/student-bot/helpers"
)

type Handler struct {
	srv       *Service
	msgSender *helpers.MsgSender
}

func NewHandler(srv *Service, msgSender *helpers.MsgSender) *Handler {
	return &Handler{srv: srv, msgSender: msgSender}
}

func (base *Handler) HandleCommand(chat *core.Chat) {
	links, err := base.srv.GetLinks(chat)
	if err != nil {
		helpers.HandleUnknownErr(base.msgSender, chat.Id, err)
		return
	}

	msg := tgbotapi.NewMessage(chat.Id, helpers.GetReply(helpers.LinkListSummaryRpl))
	rows := make([][]tgbotapi.InlineKeyboardButton, len(links))
	if len(links) == 0 {
		msg = tgbotapi.NewMessage(chat.Id, helpers.GetReply(helpers.LinkEmptyListRpl))
		base.msgSender.SendMsg(&msg)
	} else {
		for i, link := range links {
			rows[i] = tgbotapi.NewInlineKeyboardRow(
				tgbotapi.NewInlineKeyboardButtonURL(link.Name, link.Url))
		}
		msg.ReplyMarkup = tgbotapi.NewInlineKeyboardMarkup(rows...)
		base.msgSender.SendMsg(&msg)
	}
}
