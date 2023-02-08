package main

import (
	"encoding/json"
	tgbotapi "github.com/go-telegram-bot-api/telegram-bot-api/v5"
	"lebot/core"
	"lebot/features/content"
	"lebot/features/greeting"
	"lebot/features/socials"
	"lebot/providers/drive"
	"lebot/providers/dynamo"
	"lebot/providers/tg"
	"log"
)

func main() {
	db, err := dynamo.NewDb()
	if err != nil {
		log.Fatal(err)
	}

	disk, err := drive.NewService()
	if err != nil {
		log.Fatal(err)
	}

	bot, err := tg.NewBotApi()
	if err != nil {
		log.Fatal(err)
	}

	updateConfig := tgbotapi.NewUpdate(0)
	updateConfig.Timeout = 60

	updates := bot.GetUpdatesChan(updateConfig)

	for update := range updates {
		if update.Message != nil {
			chat, err := core.GetChat(db, update.Message.Chat.ID)
			if err != nil {
				tg.SendFatalErr(bot, update.Message.Chat.ID, err)
				break
			}

			switch update.Message.Text {
			case "/start":
				chat, err = greeting.CreateChat(db, update.Message.Chat.ID)
				if err != nil {
					tg.SendFatalErr(bot, chat.Id, err)
					break
				}

				greetingText := greeting.GetGreeting(chat)
				msg := tgbotapi.NewMessage(update.Message.Chat.ID, greetingText)
				tg.SendMsg(bot, msg)
				continue
			case "/files":
				files, err := content.GetFiles(disk, chat)
				if err != nil {
					tg.SendFatalErr(bot, chat.Id, err)
					break
				}

				msg := tgbotapi.NewMessage(update.Message.Chat.ID, "Files:")
				rows := make([][]tgbotapi.InlineKeyboardButton, len(files))
				if len(files) == 0 {
					msg = tgbotapi.NewMessage(update.Message.Chat.ID, "No files")
				} else {
					for i, file := range files {
						event, err := json.Marshal(content.FileEvent{Type: "DownloadFile", FileId: file.Id})
						if err != nil {
							tg.SendFatalErr(bot, chat.Id, err)
						}
						rows[i] = tgbotapi.NewInlineKeyboardRow(tgbotapi.NewInlineKeyboardButtonData(file.Name, string(event)))
					}
				}
				msg.ReplyMarkup = tgbotapi.NewInlineKeyboardMarkup(rows...)
				tg.SendMsg(bot, msg)
				continue
			case "/links":
				links, err := socials.GetLinks(disk, chat)
				if err != nil {
					tg.SendFatalErr(bot, chat.Id, err)
					break
				}

				msg := tgbotapi.NewMessage(chat.Id, "Links:")
				rows := make([][]tgbotapi.InlineKeyboardButton, len(links))
				if len(links) == 0 {
					msg = tgbotapi.NewMessage(chat.Id, "No links :(")
				} else {
					for i, link := range links {
						rows[i] = tgbotapi.NewInlineKeyboardRow(
							tgbotapi.NewInlineKeyboardButtonURL(link.Name, link.Url))
					}
				}
				msg.ReplyMarkup = tgbotapi.NewInlineKeyboardMarkup(rows...)
				tg.SendMsg(bot, msg)
				continue
			}

			switch chat.State {
			case core.Start:
				err := greeting.SaveTeacherEmail(db, chat, update.Message.Text)
				if err != nil {
					tg.SendFatalErr(bot, chat.Id, err)
					break
				}

				var msg tgbotapi.MessageConfig
				if err == greeting.ErrInvalidEmail {
					msg = tgbotapi.NewMessage(update.Message.Chat.ID, "Enter valid teacher gmail")
				} else if err != nil {
					msg = tgbotapi.NewMessage(update.Message.Chat.ID, "Error")
				} else {
					msg = tgbotapi.NewMessage(update.Message.Chat.ID, "Ready to use")
				}
				tg.SendMsg(bot, msg)
				continue
			}
		} else if update.CallbackQuery != nil {
			chat, err := core.GetChat(db, update.CallbackQuery.Message.Chat.ID)
			if err != nil {
				tg.SendFatalErr(bot, update.Message.Chat.ID, err)
				break
			}

			var event *core.Event
			err = json.Unmarshal([]byte(update.CallbackQuery.Data), event)
			if err != nil {
				tg.SendFatalErr(bot, chat.Id, err)
			}

			switch event.Type {
			case content.GetFileEvent:
				var getFileEvent *content.FileEvent
				err = json.Unmarshal([]byte(update.CallbackQuery.Data), getFileEvent)

				fileMeta, err := content.GetFileMeta(disk, getFileEvent.FileId)
				if err != nil {
					tg.SendFatalErr(bot, chat.Id, err)
				}

				const maxFileSize = 5000000
				if fileMeta.Size <= maxFileSize {
					fileContent, err := content.GetFileContent(disk, fileMeta.Id)
					if err != nil {
						tg.SendFatalErr(bot, chat.Id, err)
					}

					doc := tgbotapi.NewDocument(chat.Id, tgbotapi.FileBytes{Name: fileMeta.Name, Bytes: fileContent})
					tg.SendDoc(bot, doc)
				} else {
					msg := tgbotapi.NewMessage(update.CallbackQuery.Message.Chat.ID, fileMeta.WebContentLink)
					tg.SendMsg(bot, msg)
				}
				continue
			}
		}
	}
}
