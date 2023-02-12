package materialfeat

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

func (base *Handler) Handle(chat *core.Chat) {
	files, err := base.srv.GetFiles(chat)
	if err != nil {
		helpers.HandleUnknownErr(base.bot, chat.Id, err)
		return
	}

	msg := tgbotapi.NewMessage(chat.Id, helpers.GetReply(helpers.ContentListSummaryRpl))
	rows := make([][]tgbotapi.InlineKeyboardButton, len(files))
	if len(files) == 0 {
		msg = tgbotapi.NewMessage(chat.Id, helpers.GetReply(helpers.ContentEmptyListRpl))
		tg.SendMsg(base.bot, msg)
	} else {
		for i, file := range files {
			eventJson, err := json.Marshal(FileEvent{Type: GetFileEvent, FileId: file.Id})
			if err != nil {
				helpers.HandleUnknownErr(base.bot, chat.Id, err)
			}
			rows[i] = tgbotapi.NewInlineKeyboardRow(
				tgbotapi.NewInlineKeyboardButtonData(file.Name, string(eventJson)))
		}
		msg.ReplyMarkup = tgbotapi.NewInlineKeyboardMarkup(rows...)
		tg.SendMsg(base.bot, msg)
	}
}

func (base *Handler) HandleGetFileEvent(chat *core.Chat, data string) {
	var getFileEvent FileEvent
	err := json.Unmarshal([]byte(data), &getFileEvent)

	fileMeta, err := base.srv.GetFileMeta(getFileEvent.FileId)
	if err != nil {
		helpers.HandleUnknownErr(base.bot, chat.Id, err)
	}

	const maxFileSize = 5000000
	if fileMeta.Size <= maxFileSize {
		fileContent, err := base.srv.GetFileContent(fileMeta.Id)
		if err != nil {
			helpers.HandleUnknownErr(base.bot, chat.Id, err)
		}

		doc := tgbotapi.NewDocument(chat.Id, tgbotapi.FileBytes{Name: fileMeta.Name, Bytes: fileContent})
		tg.SendDoc(base.bot, doc)
	} else {
		msg := tgbotapi.NewMessage(chat.Id, fileMeta.WebContentLink)
		tg.SendMsg(base.bot, msg)
	}
}
