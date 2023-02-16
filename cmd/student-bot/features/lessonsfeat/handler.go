package lessonsfeat

import (
	"encoding/json"
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

func (base *Handler) HandleNewChat(chat *core.Chat) {
	err := base.srv.InitNewChat(chat)
	if err != nil {
		helpers.HandleUnknownErr(base.bot, chat.Id, err)
	}
}

func (base *Handler) Handle(chat *core.Chat) {
	lessons, err := base.srv.GetLessons(chat)
	if err != nil {
		helpers.HandleUnknownErr(base.bot, chat.Id, err)
		return
	}

	msg := tgbotapi.NewMessage(chat.Id, helpers.GetReply(helpers.ContentListSummaryRpl))
	rows := make([][]tgbotapi.InlineKeyboardButton, len(lessons))
	if len(lessons) == 0 {
		msg = tgbotapi.NewMessage(chat.Id, helpers.GetReply(helpers.ContentEmptyListRpl))
		tg.SendMsg(base.bot, msg)
	} else {
		for i, lesson := range lessons {
			eventJson, err := json.Marshal(core.ButtonEvent{Type: core.GetLessonEvent, Value: lesson.Id})
			if err != nil {
				helpers.HandleUnknownErr(base.bot, chat.Id, err)
			}
			rows[i] = tgbotapi.NewInlineKeyboardRow(
				tgbotapi.NewInlineKeyboardButtonData(lesson.Name, string(eventJson)))
		}
		msg.ReplyMarkup = tgbotapi.NewInlineKeyboardMarkup(rows...)
		tg.SendMsg(base.bot, msg)
	}
}

func (base *Handler) HandleGetLessonEvent(chat *core.Chat, data string) {
	var getLessonEvent core.ButtonEvent
	err := json.Unmarshal([]byte(data), &getLessonEvent)

	lessonContent, err := base.srv.GetLessonContent(getLessonEvent.Value)
	if err != nil {
		helpers.HandleUnknownErr(base.bot, chat.Id, err)
	}

	doc := tgbotapi.NewDocument(chat.Id, tgbotapi.FileBytes{Name: "Lesson.txt", Bytes: lessonContent})
	tg.SendDoc(base.bot, doc)
}
