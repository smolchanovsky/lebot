package tg

import (
	tgbotapi "github.com/go-telegram-bot-api/telegram-bot-api/v5"
	"lebot/internal/secret"
	"log"
)

func NewBotApi() (*tgbotapi.BotAPI, error) {
	tgmToken, err := secret.GetSecret(secret.TgmTokenPath)
	if err != nil {
		return nil, err
	}

	bot, err := tgbotapi.NewBotAPI(tgmToken)
	return bot, err
}

func SendMsg(bot *tgbotapi.BotAPI, msg tgbotapi.MessageConfig) {
	if _, err := bot.Send(msg); err != nil {
		log.Print("unable to send message: ", err)
	} else {
		log.Printf("reply sent: %s", msg.Text)
	}
}

func SendDoc(bot *tgbotapi.BotAPI, doc tgbotapi.DocumentConfig) {
	if _, err := bot.Send(doc); err != nil {
		log.Print("unable to send doc: ", err)
	} else {
		log.Printf("document sent")
	}
}

func SendText(bot *tgbotapi.BotAPI, chatId int64, text string) {
	msg := tgbotapi.NewMessage(chatId, text)
	msg.ParseMode = tgbotapi.ModeMarkdown
	SendMsg(bot, msg)
}
