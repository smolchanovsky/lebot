package joinfeat

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

func (base *Handler) HandleStart(tgChat *tgbotapi.Chat) {
	chat, err := base.srv.SaveChat(tgChat.ID, tgChat.UserName)
	if err != nil {
		helpers.HandleUnknownErr(base.bot, tgChat.ID, err)
		return
	}

	msg := tgbotapi.NewMessage(chat.Id, helpers.GetReply(helpers.JoinStartRpl))
	tg.SendMsg(base.bot, msg)
}

func (base *Handler) HandleEmail(chat *core.Chat, email string) {
	err := base.srv.SaveTeacherEmail(chat, email)

	var msg tgbotapi.MessageConfig
	if err == ErrInvalidEmail {
		msg = tgbotapi.NewMessage(chat.Id, helpers.GetReply(helpers.JoinInvalidEmailRpl))
	} else if err == ErrEmailNotFound {
		msg = tgbotapi.NewMessage(chat.Id, helpers.GetReply(helpers.JoinEmailNotFoundRpl))
	} else if err != nil {
		helpers.HandleUnknownErr(base.bot, chat.Id, err)
	} else {
		msg = tgbotapi.NewMessage(chat.Id, helpers.GetReply(helpers.JoinFinishRpl))
	}

	tg.SendMsg(base.bot, msg)
}
