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

	log.Print("listening updates...")
	updates := bot.GetUpdatesChan(updateConfig)

	for update := range updates {
		log.Printf("new update '%d'", update.UpdateID)

		if update.Message != nil {
			chatId := update.Message.Chat.ID
			text := update.Message.Text
			log.Printf("new message in '%d' chat: %s", chatId, text)

			chat, err := core.GetChat(db, chatId)
			if err != nil {
				tg.SendFatalErr(bot, chatId, err)
				continue
			}

			switch text {
			case "/start":
				chat, err = greeting.CreateChat(db, chat.Id)
				if err != nil {
					tg.SendFatalErr(bot, chat.Id, err)
					continue
				}

				greetingText := greeting.GetGreeting(chat)
				msg := tgbotapi.NewMessage(chat.Id, greetingText)
				tg.SendMsg(bot, msg)
				continue
			case "/files":
				files, err := content.GetFiles(disk, chat)
				if err != nil {
					tg.SendFatalErr(bot, chat.Id, err)
					continue
				}

				msg := tgbotapi.NewMessage(chat.Id, "Files:")
				rows := make([][]tgbotapi.InlineKeyboardButton, len(files))
				if len(files) == 0 {
					msg = tgbotapi.NewMessage(chat.Id, "No files")
				} else {
					for i, file := range files {
						eventJson, err := json.Marshal(content.FileEvent{Type: content.GetFileEvent, FileId: file.Id})
						if err != nil {
							tg.SendFatalErr(bot, chat.Id, err)
						}
						rows[i] = tgbotapi.NewInlineKeyboardRow(
							tgbotapi.NewInlineKeyboardButtonData(file.Name, string(eventJson)))
					}
				}
				msg.ReplyMarkup = tgbotapi.NewInlineKeyboardMarkup(rows...)
				tg.SendMsg(bot, msg)
				continue
			case "/links":
				links, err := socials.GetLinks(disk, chat)
				if err != nil {
					tg.SendFatalErr(bot, chat.Id, err)
					continue
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
				err := greeting.SaveTeacherEmail(db, chat, text)
				if err != nil {
					tg.SendFatalErr(bot, chat.Id, err)
					continue
				}

				var msg tgbotapi.MessageConfig
				if err == greeting.ErrInvalidEmail {
					msg = tgbotapi.NewMessage(chat.Id, "Enter valid teacher gmail")
				} else if err != nil {
					msg = tgbotapi.NewMessage(chat.Id, "Error")
				} else {
					msg = tgbotapi.NewMessage(chat.Id, "Ready to use")
				}
				tg.SendMsg(bot, msg)
				continue
			}
		} else if update.CallbackQuery != nil {
			chatId := update.CallbackQuery.Message.Chat.ID
			data := update.CallbackQuery.Data
			log.Printf("new callback in '%d' chat: %s", chatId, data)

			chat, err := core.GetChat(db, chatId)
			if err != nil {
				tg.SendFatalErr(bot, chatId, err)
				continue
			}

			var event core.Event
			err = json.Unmarshal([]byte(data), &event)
			if err != nil {
				tg.SendFatalErr(bot, chat.Id, err)
			}
			log.Printf("callback is '%s' event", event.Type)

			switch event.Type {
			case content.GetFileEvent:
				var getFileEvent content.FileEvent
				err = json.Unmarshal([]byte(data), &getFileEvent)

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
					msg := tgbotapi.NewMessage(chat.Id, fileMeta.WebContentLink)
					tg.SendMsg(bot, msg)
				}
				continue
			}
		}
	}
}
