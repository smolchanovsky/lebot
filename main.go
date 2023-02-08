package main

import (
	"awesomeProject/core"
	"awesomeProject/features/content"
	"awesomeProject/features/greeting"
	"awesomeProject/features/socials"
	"awesomeProject/providers"
	"awesomeProject/providers/dynamo"
	"awesomeProject/providers/tg"
	"encoding/json"
	tgbotapi "github.com/go-telegram-tg-api/telegram-tg-api/v5"
	"log"
)

func main() {
	db, err := dynamo.GetDb()
	if err != nil {
		log.Fatal(err)
	}

	drive, err := providers.GetDriveService()
	if err != nil {
		log.Fatal(err)
	}

	botApi, err := tg.GetTgBot()
	if err != nil {
		log.Fatal(err)
	}

	updateConfig := tgbotapi.NewUpdate(0)
	updateConfig.Timeout = 60

	updates := botApi.GetUpdatesChan(updateConfig)

	for update := range updates {
		if update.Message != nil {
			chat, err := core.GetChat(db, update.Message.Chat.ID)
			if err != nil {
				tg.SendFatalErr(botApi, update.Message.Chat.ID, err)
				break
			}

			switch update.Message.Text {
			case "/start":
				chat, err = greeting.CreateChat(db, update.Message.Chat.ID)
				if err != nil {
					tg.SendFatalErr(botApi, chat.Id, err)
					break
				}

				greetingText := greeting.GetGreeting(chat)
				msg := tgbotapi.NewMessage(update.Message.Chat.ID, greetingText)
				tg.SendMsg(botApi, msg)
				continue
			case "/files":
				files, err := content.GetFiles(drive, chat)
				if err != nil {
					tg.SendFatalErr(botApi, chat.Id, err)
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
							tg.SendFatalErr(botApi, chat.Id, err)
						}
						rows[i] = tgbotapi.NewInlineKeyboardRow(tgbotapi.NewInlineKeyboardButtonData(file.Name, string(event)))
					}
				}
				msg.ReplyMarkup = tgbotapi.NewInlineKeyboardMarkup(rows...)
				tg.SendMsg(botApi, msg)
				continue
			case "/links":
				links, err := socials.GetLinks(drive, chat)
				if err != nil {
					tg.SendFatalErr(botApi, chat.Id, err)
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
				tg.SendMsg(botApi, msg)
				continue
			}

			switch chat.State {
			case core.Start:
				err := greeting.SaveTeacherEmail(db, chat, update.Message.Text)
				if err != nil {
					tg.SendFatalErr(botApi, chat.Id, err)
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
				tg.SendMsg(botApi, msg)
				continue
			}
		} else if update.CallbackQuery != nil {
			chat, err := core.GetChat(db, update.CallbackQuery.Message.Chat.ID)
			if err != nil {
				tg.SendFatalErr(botApi, update.Message.Chat.ID, err)
				break
			}

			var event *core.Event
			err = json.Unmarshal([]byte(update.CallbackQuery.Data), event)
			if err != nil {
				tg.SendFatalErr(botApi, chat.Id, err)
			}

			switch event.Type {
			case content.GetFileEvent:
				var getFileEvent *content.FileEvent
				err = json.Unmarshal([]byte(update.CallbackQuery.Data), getFileEvent)

				fileMeta, err := content.GetFileMeta(drive, getFileEvent.FileId)
				if err != nil {
					tg.SendFatalErr(botApi, chat.Id, err)
				}

				const maxFileSize = 5000000
				if fileMeta.Size <= maxFileSize {
					fileContent, err := content.GetFileContent(drive, fileMeta.Id)
					if err != nil {
						tg.SendFatalErr(botApi, chat.Id, err)
					}

					doc := tgbotapi.NewDocument(chat.Id, tgbotapi.FileBytes{Name: fileMeta.Name, Bytes: fileContent})
					tg.SendDoc(botApi, doc)
				} else {
					msg := tgbotapi.NewMessage(update.CallbackQuery.Message.Chat.ID, fileMeta.WebContentLink)
					tg.SendMsg(botApi, msg)
				}
				continue
			}
		}
	}
}
