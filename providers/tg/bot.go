package tg

import (
	"awesomeProject/secrets"
	tgbotapi "github.com/go-telegram-bot-api/telegram-bot-api/v5"
	"log"
)

func GetTgBot() (*tgbotapi.BotAPI, error) {
	tgmToken, err := secrets.GetSecret(secrets.TgmTokenPath)
	if err != nil {
		return nil, err
	}

	bot, err := tgbotapi.NewBotAPI(tgmToken)
	return bot, err
}

func SendMsg(bot *tgbotapi.BotAPI, msg tgbotapi.MessageConfig) {
	if _, err := bot.Send(msg); err != nil {
		log.Print("", err)
	}
}

func SendDoc(bot *tgbotapi.BotAPI, msg tgbotapi.DocumentConfig) {
	if _, err := bot.Send(msg); err != nil {
		log.Print("", err)
	}
}

func SendText(bot *tgbotapi.BotAPI, chatId int64, text string) {
	msg := tgbotapi.NewMessage(chatId, text)
	SendMsg(bot, msg)
}

func SendFatalErr(bot *tgbotapi.BotAPI, chatId int64, err error) {
	log.Print("unexpected error", err)
	SendText(bot, chatId, "Unexpected error")
}
