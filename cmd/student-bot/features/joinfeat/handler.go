package joinfeat

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

func (base *Handler) HandleCommand(tgChat *tgbotapi.Chat) {
	chat, err := base.srv.SaveChat(tgChat.ID, tgChat.UserName)
	if err != nil {
		helpers.HandleUnknownErr(base.msgSender, tgChat.ID, err)
		return
	}

	base.msgSender.SendText(chat.Id, helpers.GetReply(helpers.JoinStartRpl))
}

func (base *Handler) HandleEmail(chat *core.Chat, email string) {
	err := base.srv.SaveTeacherEmail(chat, email)

	var msg tgbotapi.MessageConfig
	if err == ErrInvalidEmail {
		msg = tgbotapi.NewMessage(chat.Id, helpers.GetReply(helpers.JoinInvalidEmailRpl))
	} else if err == ErrEmailNotFound {
		msg = tgbotapi.NewMessage(chat.Id, helpers.GetReply(helpers.JoinEmailNotFoundRpl))
	} else if err != nil {
		helpers.HandleUnknownErr(base.msgSender, chat.Id, err)
	} else {
		msg = tgbotapi.NewMessage(chat.Id, helpers.GetReply(helpers.JoinFinishRpl))
	}

	base.msgSender.SendMsg(&msg)
}
