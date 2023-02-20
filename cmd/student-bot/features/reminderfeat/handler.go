package reminderfeat

import (
	"fmt"
	"lebot/cmd/student-bot/core"
	"lebot/cmd/student-bot/helpers"
	"log"
)

type Handler struct {
	srv       *Service
	msgSender *helpers.MsgSender
}

func NewHandler(srv *Service, msgSender *helpers.MsgSender) *Handler {
	return &Handler{srv: srv, msgSender: msgSender}
}

func (base *Handler) HandleNewChat(chat *core.Chat) {
	err := base.srv.InitNewChat(chat)
	if err != nil {
		helpers.HandleUnknownErr(base.msgSender, chat.Id, err)
	}
}

func (base *Handler) HandleLessonsSoon() {
	reminders, err := base.srv.GetLessonsSoon()
	if err != nil {
		log.Print("error when obtaining reminders: ", err)
		return
	}

	for _, reminder := range reminders {
		text := helpers.GetReply(helpers.ReminderLessonSoonRpl)
		base.msgSender.SendText(reminder.ChatId, text)
	}
}

func (base *Handler) HandleLessonsStart() {
	reminders, err := base.srv.GetLessonsStart()
	if err != nil {
		log.Print("error when obtaining reminders: ", err)
		return
	}

	for _, reminder := range reminders {
		reply := fmt.Sprintf("%s", helpers.GetReply(helpers.ReminderLessonStartRpl))
		if reminder.Url == nil {
			reply = reply + fmt.Sprintf("\n%s", *reminder.Url)
		}
		base.msgSender.SendText(reminder.ChatId, reply)
	}
}
