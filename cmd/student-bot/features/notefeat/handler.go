package notefeat

import (
	"encoding/json"
	tgbotapi "github.com/go-telegram-bot-api/telegram-bot-api/v5"
	"lebot/cmd/student-bot/core"
	"lebot/cmd/student-bot/helpers"
	"lebot/internal/tg"
)

const GetNoteEvent = 2

type NoteEvent struct {
	Type   int    `json:"t"`
	NoteId string `json:"n"`
}

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
	notes, err := base.srv.GetNotes(chat)
	if err != nil {
		helpers.HandleUnknownErr(base.bot, chat.Id, err)
		return
	}

	msg := tgbotapi.NewMessage(chat.Id, helpers.GetReply(helpers.ContentListSummaryRpl))
	rows := make([][]tgbotapi.InlineKeyboardButton, len(notes))
	if len(notes) == 0 {
		msg = tgbotapi.NewMessage(chat.Id, helpers.GetReply(helpers.ContentEmptyListRpl))
		tg.SendMsg(base.bot, msg)
	} else {
		for i, note := range notes {
			eventJson, err := json.Marshal(NoteEvent{Type: GetNoteEvent, NoteId: note.Id})
			if err != nil {
				helpers.HandleUnknownErr(base.bot, chat.Id, err)
			}
			rows[i] = tgbotapi.NewInlineKeyboardRow(
				tgbotapi.NewInlineKeyboardButtonData(note.Name, string(eventJson)))
		}
		msg.ReplyMarkup = tgbotapi.NewInlineKeyboardMarkup(rows...)
		tg.SendMsg(base.bot, msg)
	}
}

func (base *Handler) HandleGetNoteEvent(chat *core.Chat, data string) {
	var getNoteEvent NoteEvent
	err := json.Unmarshal([]byte(data), &getNoteEvent)

	noteContent, err := base.srv.GetNoteContent(getNoteEvent.NoteId)
	if err != nil {
		helpers.HandleUnknownErr(base.bot, chat.Id, err)
	}

	doc := tgbotapi.NewDocument(chat.Id, tgbotapi.FileBytes{Name: "Lesson note.txt", Bytes: noteContent})
	tg.SendDoc(base.bot, doc)
}
