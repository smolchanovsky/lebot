package reminderfeat

import (
	"fmt"
	tgbotapi "github.com/go-telegram-bot-api/telegram-bot-api/v5"
	"lebot/cmd/student-bot/core"
	"lebot/cmd/student-bot/helpers"
	"lebot/internal/tg"
	"log"
)

type Handler struct {
	srv *Service
	bot *tgbotapi.BotAPI
}

func NewHandler(srv *Service, bot *tgbotapi.BotAPI) *Handler {
	return &Handler{srv: srv, bot: bot}
}

func (base *Handler) HandleLessonsSoon() {
	reminders, err := base.srv.GetLessonsSoon()
	if err != nil {
		log.Print("error when obtaining reminders: ", err)
		return
	}

	for _, reminder := range reminders {
		text := helpers.GetReply(helpers.ReminderLessonSoonRpl)
		tg.SendText(base.bot, reminder.ChatId, text)
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
		tg.SendText(base.bot, reminder.ChatId, reply)
	}
}

func (base *Handler) HandleNewChat(chat *core.Chat) {
	err := base.srv.InitNewChat(chat)
	if err != nil {
		helpers.HandleUnknownErr(base.bot, chat.Id, err)
	}
}
