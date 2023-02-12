package reminderfeat

import (
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
		log.Print("error when obtaining reminders", err)
		return
	}

	for _, reminder := range reminders {
		text := helpers.GetReply(helpers.ReminderLessonSoonRpl)
		tg.SendText(base.bot, reminder.ChatId, text)
	}
}

func (base *Handler) HandleNewChat(chat *core.Chat) {
	base.srv.InitNewChat(chat)
}
