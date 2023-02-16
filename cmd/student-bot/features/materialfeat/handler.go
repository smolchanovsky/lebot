package materialfeat

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

func (base *Handler) Handle(chat *core.Chat) {
	materialsFolder, err := base.srv.GetMaterialsFolder(chat)
	if err != nil {
		helpers.HandleUnknownErr(base.bot, chat.Id, err)
		return
	}

	base.renderFolder(chat, materialsFolder.Id)
}

func (base *Handler) HandleButtonEvent(chat *core.Chat, data string) {
	var event *core.ButtonEvent
	err := json.Unmarshal([]byte(data), &event)
	if err != nil {
		helpers.HandleUnknownErr(base.bot, chat.Id, err)
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
	materials, err := base.srv.GetMaterials(id)
	if err != nil {
		helpers.HandleUnknownErr(base.bot, chat.Id, err)
		return
	}

	msg := tgbotapi.NewMessage(chat.Id, helpers.GetReply(helpers.ContentListSummaryRpl))
	rows := make([][]tgbotapi.InlineKeyboardButton, len(materials))
	if len(materials) == 0 {
		msg = tgbotapi.NewMessage(chat.Id, helpers.GetReply(helpers.ContentEmptyListRpl))
		tg.SendMsg(base.bot, msg)
	} else {
		for i, material := range materials {
			var text string
			var event *core.ButtonEvent
			if material.MimeType == "application/vnd.google-apps.folder" {
				text = fmt.Sprintf("🗂%s", material.Name)
				event = &core.ButtonEvent{
					Type:   core.MaterialEvent,
					Action: core.GetFolderAction,
					Value:  material.Id,
				}
			} else {
				text = fmt.Sprintf("📄%s", material.Name)
				event = &core.ButtonEvent{
					Type:   core.MaterialEvent,
					Action: core.GetFileAction,
					Value:  material.Id,
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
	materialMeta, err := base.srv.GetMaterialMeta(id)
	if err != nil {
		helpers.HandleUnknownErr(base.bot, chat.Id, err)
	}

	const maxMaterialSize = 5000000
	if materialMeta.Size <= maxMaterialSize {
		materialContent, err := base.srv.GetMaterialContent(materialMeta.Id)
		if err != nil {
			helpers.HandleUnknownErr(base.bot, chat.Id, err)
		}

		doc := tgbotapi.NewDocument(chat.Id, tgbotapi.FileBytes{Name: materialMeta.Name, Bytes: materialContent})
		tg.SendDoc(base.bot, doc)
	} else {
		msg := tgbotapi.NewMessage(chat.Id, materialMeta.WebContentLink)
		tg.SendMsg(base.bot, msg)
	}
}
