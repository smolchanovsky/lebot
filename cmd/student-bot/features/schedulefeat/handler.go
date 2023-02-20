package schedulefeat

import (
	"fmt"
	"lebot/cmd/student-bot/core"
	"lebot/cmd/student-bot/helpers"
	"log"
	"time"
)

type Handler struct {
	srv       *Service
	msgSender *helpers.MsgSender
}

func NewHandler(srv *Service, msgSender *helpers.MsgSender) *Handler {
	return &Handler{srv: srv, msgSender: msgSender}
}

func (base *Handler) Handle(chat *core.Chat) {
	const count = 5
	lessons, err := base.srv.GetLessons(chat, count)
	if err != nil {
		log.Print("error when obtaining lessons: ", err)
		return
	}

	var text string

	if len(lessons) == 0 {
		text = helpers.GetReply(helpers.ScheduleNoLessonsRpl)
	} else {
		text = fmt.Sprint("Your next lessons:\n")
		for _, lesson := range lessons {
			line := fmt.Sprintf(
				"🗓️*%s:* %s - %s",
				lesson.start.Format("Jan-02 Mon"),
				lesson.start.Format(time.Kitchen),
				lesson.end.Format(time.Kitchen))
			text += line + "\n"
		}

		if len(lessons) < count {
			text += fmt.Sprintf("\nThere are only %d lessons in the schedule", len(lessons))
		}
	}
	base.msgSender.SendText(chat.Id, text)
}
