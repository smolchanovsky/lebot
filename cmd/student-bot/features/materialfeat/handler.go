package materialfeat

import (
	"encoding/json"
	tgbotapi "github.com/go-telegram-bot-api/telegram-bot-api/v5"
	"lebot/cmd/student-bot/core"
	"lebot/cmd/student-bot/helpers"
	"lebot/internal/tg"
)

const GetMaterialEvent = 1

type MaterialEvent struct {
	Type       int
	MaterialId string
}

type Handler struct {
	srv *Service
	bot *tgbotapi.BotAPI
}

func NewHandler(srv *Service, bot *tgbotapi.BotAPI) *Handler {
	return &Handler{srv: srv, bot: bot}
}

func (base *Handler) Handle(chat *core.Chat) {
	materials, err := base.srv.GetMaterials(chat)
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
			eventJson, err := json.Marshal(MaterialEvent{Type: GetMaterialEvent, MaterialId: material.Id})
			if err != nil {
				helpers.HandleUnknownErr(base.bot, chat.Id, err)
			}
			rows[i] = tgbotapi.NewInlineKeyboardRow(
				tgbotapi.NewInlineKeyboardButtonData(material.Name, string(eventJson)))
		}
		msg.ReplyMarkup = tgbotapi.NewInlineKeyboardMarkup(rows...)
		tg.SendMsg(base.bot, msg)
	}
}

func (base *Handler) HandleGetMaterialEvent(chat *core.Chat, data string) {
	var getMaterialEvent MaterialEvent
	err := json.Unmarshal([]byte(data), &getMaterialEvent)

	materialMeta, err := base.srv.GetMaterialMeta(getMaterialEvent.MaterialId)
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
