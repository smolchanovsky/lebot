package linkfeat

import (
	tgbotapi "github.com/go-telegram-bot-api/telegram-bot-api/v5"
	"lebot/cmd/student-bot/core"
	"lebot/cmd/student-bot/helpers"
	"lebot/internal/tg"
)

type Handler struct {
	srv *Service
	bot *tgbotapi.BotAPI
}

func NewHandler(srv *Service, bot *tgbotapi.BotAPI) *Handler {
	return &Handler{srv: srv, bot: bot}
}

func (base *Handler) Handle(chat *core.Chat) {
	links, err := base.srv.GetLinks(chat)
	if err != nil {
		helpers.HandleUnknownErr(base.bot, chat.Id, err)
		return
	}

	msg := tgbotapi.NewMessage(chat.Id, helpers.GetReply(helpers.LinkListSummaryRpl))
	rows := make([][]tgbotapi.InlineKeyboardButton, len(links))
	if len(links) == 0 {
		msg = tgbotapi.NewMessage(chat.Id, helpers.GetReply(helpers.LinkEmptyListRpl))
		tg.SendMsg(base.bot, msg)
	} else {
		for i, link := range links {
			rows[i] = tgbotapi.NewInlineKeyboardRow(
				tgbotapi.NewInlineKeyboardButtonURL(link.Name, link.Url))
		}
		msg.ReplyMarkup = tgbotapi.NewInlineKeyboardMarkup(rows...)
		tg.SendMsg(base.bot, msg)
	}
}
