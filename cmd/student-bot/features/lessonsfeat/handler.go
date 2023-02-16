package lessonsfeat

import (
	"encoding/json"
	"fmt"
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

func (base *Handler) HandleCommand(chat *core.Chat) {
	lessonsFolder, err := base.srv.GetLessonsRoot(chat)
	if err != nil {
		helpers.HandleUnknownErr(base.bot, chat.Id, err)
		return
	}

	base.renderFolder(chat, lessonsFolder.Id)
}

func (base *Handler) HandleButtonEvent(chat *core.Chat, data string) {
	var event core.ButtonEvent
	err := json.Unmarshal([]byte(data), &event)
	if err != nil {
		helpers.HandleUnknownErr(base.bot, chat.Id, err)
		return
	}

	switch event.Action {
	case core.GetFolderAction:
		base.renderFolder(chat, event.Value)
		break
	case core.GetFileAction:
		base.renderFile(chat, event.Value)
		break
	}
}

func (base *Handler) renderFolder(chat *core.Chat, id string) {
	lessons, err := base.srv.GetLessonFolders(id)
	if err != nil {
		helpers.HandleUnknownErr(base.bot, chat.Id, err)
		return
	}

	// Copy from materials
	msg := tgbotapi.NewMessage(chat.Id, helpers.GetReply(helpers.ContentListSummaryRpl))
	rows := make([][]tgbotapi.InlineKeyboardButton, len(lessons))
	if len(lessons) == 0 {
		msg = tgbotapi.NewMessage(chat.Id, helpers.GetReply(helpers.ContentEmptyListRpl))
		tg.SendMsg(base.bot, msg)
	} else {
		for i, lesson := range lessons {
			var text string
			var event *core.ButtonEvent
			if lesson.MimeType == "application/vnd.google-apps.folder" {
				text = fmt.Sprintf("🗂%s", lesson.Name)
				event = &core.ButtonEvent{
					Type:   core.LessonEvent,
					Action: core.GetFolderAction,
					Value:  lesson.Id,
				}
			} else {
				text = fmt.Sprintf("📄%s", lesson.Name)
				event = &core.ButtonEvent{
					Type:   core.LessonEvent,
					Action: core.GetFileAction,
					Value:  lesson.Id,
				}
			}
			eventJson, err := json.Marshal(event)
			if err != nil {
				helpers.HandleUnknownErr(base.bot, chat.Id, err)
			}
			rows[i] = tgbotapi.NewInlineKeyboardRow(
				tgbotapi.NewInlineKeyboardButtonData(text, string(eventJson)))
		}
		msg.ReplyMarkup = tgbotapi.NewInlineKeyboardMarkup(rows...)
		tg.SendMsg(base.bot, msg)
	}
}

func (base *Handler) renderFile(chat *core.Chat, id string) {
	materialMeta, err := base.srv.GetLessonFileMeta(id)
	if err != nil {
		helpers.HandleUnknownErr(base.bot, chat.Id, err)
	}

	// Copy from materials
	const maxMaterialSize = 5000000
	if materialMeta.Size <= maxMaterialSize {
		materialContent, err := base.srv.GetLessonFileContent(materialMeta.Id)
		if err != nil {
			helpers.HandleUnknownErr(base.bot, chat.Id, err)
		}

		doc := tgbotapi.NewDocument(chat.Id, tgbotapi.FileBytes{
			Name:  fmt.Sprintf("%s.%s", materialMeta.Name, "txt"), // TODO: Fix extension
			Bytes: materialContent,
		})
		tg.SendDoc(base.bot, doc)
	} else {
		msg := tgbotapi.NewMessage(chat.Id, materialMeta.WebContentLink)
		tg.SendMsg(base.bot, msg)
	}
}
