package materialfeat

import (
	"encoding/json"
	"fmt"
	tgbotapi "github.com/go-telegram-bot-api/telegram-bot-api/v5"
	"lebot/cmd/student-bot/core"
	"lebot/cmd/student-bot/helpers"
)

type Handler struct {
	srv       *Service
	msgSender *helpers.MsgSender
}

func NewHandler(srv *Service, msgSender *helpers.MsgSender) *Handler {
	return &Handler{srv: srv, msgSender: msgSender}
}

func (base *Handler) HandleCommand(chat *core.Chat) {
	materialsFolder, err := base.srv.GetMaterialsRoot(chat)
	if err != nil {
		helpers.HandleUnknownErr(base.msgSender, chat.Id, err)
		return
	}

	base.renderFolder(chat, materialsFolder.Id)
}

func (base *Handler) HandleButtonEvent(chat *core.Chat, data string) {
	var event *core.ButtonEvent
	err := json.Unmarshal([]byte(data), &event)
	if err != nil {
		helpers.HandleUnknownErr(base.msgSender, chat.Id, err)
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
		helpers.HandleUnknownErr(base.msgSender, chat.Id, err)
		return
	}

	msg := tgbotapi.NewMessage(chat.Id, helpers.GetReply(helpers.ContentListSummaryRpl))
	rows := make([][]tgbotapi.InlineKeyboardButton, len(materials))
	if len(materials) == 0 {
		msg = tgbotapi.NewMessage(chat.Id, helpers.GetReply(helpers.ContentEmptyListRpl))
		base.msgSender.SendMsg(&msg)
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
				helpers.HandleUnknownErr(base.msgSender, chat.Id, err)
			}
			rows[i] = tgbotapi.NewInlineKeyboardRow(
				tgbotapi.NewInlineKeyboardButtonData(text, string(eventJson)))
		}
		msg.ReplyMarkup = tgbotapi.NewInlineKeyboardMarkup(rows...)
		base.msgSender.SendMsg(&msg)
	}
}

func (base *Handler) renderFile(chat *core.Chat, id string) {
	materialMeta, err := base.srv.GetMaterialMeta(id)
	if err != nil {
		helpers.HandleUnknownErr(base.msgSender, chat.Id, err)
	}

	const maxMaterialSize = 5000000
	if materialMeta.Size <= maxMaterialSize {
		materialContent, err := base.srv.GetMaterialContent(materialMeta.Id)
		if err != nil {
			helpers.HandleUnknownErr(base.msgSender, chat.Id, err)
		}

		doc := tgbotapi.NewDocument(chat.Id, tgbotapi.FileBytes{
			Name:  materialMeta.Name,
			Bytes: materialContent,
		})
		base.msgSender.SendDoc(&doc)
	} else {
		msg := tgbotapi.NewMessage(chat.Id, materialMeta.WebContentLink)
		base.msgSender.SendMsg(&msg)
	}
}
