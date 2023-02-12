package helpers

import (
	tgbotapi "github.com/go-telegram-bot-api/telegram-bot-api/v5"
	"lebot/internal/tg"
	"log"
)

func HandleUnknownErr(bot *tgbotapi.BotAPI, chatId int64, err error) {
	log.Print("unknown error", err)
	text := GetReply(ErrorUnknownRpl)
	tg.SendText(bot, chatId, text)
}
