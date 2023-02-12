package schedulefeat

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

func (base *Handler) Handle(chat *core.Chat) {
	const count = 10
	lessons, err := base.srv.GetLessons(chat, count)
	if err != nil {
		log.Print("error when obtaining lessons", err)
		return
	}

	var text string

	if len(lessons) == 0 {
		text = helpers.GetReply(helpers.ScheduleNoLessonsRpl)
	} else {
		text = fmt.Sprintf("Next %d lessons:\n", count)
		for _, lesson := range lessons {
			line := fmt.Sprintf(
				"*%s %s:* %s - %s",
				lesson.start.Format("Jan-02"),
				lesson.start.Weekday(),
				lesson.start.Format("10:50"),
				lesson.end.Format("10:50"))
			text += line + "\n"
		}

		if len(lessons) < count {
			text += fmt.Sprintf("\nThere are only %d lessons in the schedule", len(lessons))
		}
	}
	tg.SendText(base.bot, chat.Id, text)
}
