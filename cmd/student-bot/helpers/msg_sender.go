package helpers

import (
	tgbotapi "github.com/go-telegram-bot-api/telegram-bot-api/v5"
	"github.com/guregu/dynamo"
	"lebot/internal/tg"
	"log"
)

type MsgSender struct {
	bot *tgbotapi.BotAPI
	db  *dynamo.DB
}

func NewMsgSender(bot *tgbotapi.BotAPI, db *dynamo.DB) *MsgSender {
	return &MsgSender{bot: bot, db: db}
}

func (base *MsgSender) SendMsg(msg *tgbotapi.MessageConfig) {
	err := SaveBotMsg(base.db, msg)
	if err != nil {
		log.Print(err)
	}
	tg.SendMsg(base.bot, msg)
}

func (base *MsgSender) SendDoc(doc *tgbotapi.DocumentConfig) {
	err := SaveBotDoc(base.db, doc)
	if err != nil {
		log.Print(err)
	}
	tg.SendDoc(base.bot, doc)
}

func (base *MsgSender) SendText(chatId int64, text string) {
	msg := tgbotapi.NewMessage(chatId, text)
	msg.ParseMode = tgbotapi.ModeMarkdown
	base.SendMsg(&msg)
}
