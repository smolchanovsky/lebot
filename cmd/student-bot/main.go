package main

import (
	"encoding/json"
	"errors"
	tgbotapi "github.com/go-telegram-bot-api/telegram-bot-api/v5"
	"github.com/robfig/cron/v3"
	"lebot/cmd/student-bot/core"
	"lebot/cmd/student-bot/features/joinfeat"
	"lebot/cmd/student-bot/features/linkfeat"
	"lebot/cmd/student-bot/features/materialfeat"
	"lebot/cmd/student-bot/features/reminderfeat"
	"lebot/cmd/student-bot/features/schedulefeat"
	"lebot/cmd/student-bot/helpers"
	"lebot/internal/dynamodb"
	"lebot/internal/googlecalendar"
	"lebot/internal/googledrive"
	"lebot/internal/tg"
	"log"
)

func main() {
	db, err := dynamodb.NewDb()
	if err != nil {
		log.Fatal(err)
	}

	diskSrv, err := googledrive.NewService()
	if err != nil {
		log.Fatal(err)
	}

	calSrv, err := googlecalendar.NewService()
	if err != nil {
		log.Fatal(err)
	}

	bot, err := tg.NewBotApi()
	if err != nil {
		log.Fatal(err)
	}

	joinSrv := joinfeat.NewService(db)
	joinHandler := joinfeat.NewHandler(joinSrv, bot)

	scheduleSrv := schedulefeat.NewService(calSrv, db)
	scheduleHandler := schedulefeat.NewHandler(scheduleSrv, bot)

	linkSrv := linkfeat.NewService(diskSrv)
	linkHandler := linkfeat.NewHandler(linkSrv, bot)

	materialSrv := materialfeat.NewService(diskSrv)
	materialHandler := materialfeat.NewHandler(materialSrv, bot)

	reminderSrv := reminderfeat.NewService(calSrv, db)
	reminderHandler := reminderfeat.NewHandler(reminderSrv, bot)

	scheduler := cron.New()
	scheduler.AddFunc("0 * * * *", func() {
		reminderHandler.HandleLessonsSoon()
	})
	scheduler.Start()

	updateConfig := tgbotapi.NewUpdate(0)
	updateConfig.Timeout = 60

	log.Print("listening updates...")
	updates := bot.GetUpdatesChan(updateConfig)

	for update := range updates {
		log.Printf("new update '%d'", update.UpdateID)

		if update.Message != nil {
			chatId := update.Message.Chat.ID
			text := update.Message.Text
			log.Printf("new reply in '%d' chat: %s", chatId, text)

			chatOrNil, err := core.GetChat(db, chatId)
			if err != nil {
				helpers.HandleUnknownErr(bot, chatId, err)
				continue
			}
			log.Printf("start processing '%d' chat with new message: %s", chatId, text)

			if len(text) > 0 && text[0] == '/' {
				HandleCommand(bot, joinHandler, scheduleHandler, linkHandler, materialHandler, update.Message, chatOrNil)
			} else {
				HandleMessage(bot, joinHandler, reminderHandler, chatOrNil, update.Message)
			}
		} else if update.CallbackQuery != nil {
			chatId := update.CallbackQuery.Message.Chat.ID
			data := update.CallbackQuery.Data
			log.Printf("new callback in '%d' chat: %s", chatId, data)

			chatOrNil, err := core.GetChat(db, chatId)
			if err != nil {
				helpers.HandleUnknownErr(bot, chatId, err)
				continue
			}
			log.Printf("start processing '%d' chat with new callback: %s", chatId, data)

			var event *core.Event
			err = json.Unmarshal([]byte(data), &event)
			if err != nil {
				helpers.HandleUnknownErr(bot, chatId, err)
			}

			HandleCallback(bot, materialHandler, chatOrNil, event, data)
		}
	}
}

func HandleCommand(
	bot *tgbotapi.BotAPI,
	join *joinfeat.Handler, scheduleHandler *schedulefeat.Handler, link *linkfeat.Handler, material *materialfeat.Handler,
	message *tgbotapi.Message, chat *core.Chat) {
	log.Printf("Try match message with one of commands")
	switch message.Text {
	case "/start":
		join.HandleStart(message.Chat)
		break
	case "/schedule":
		scheduleHandler.Handle(chat)
		break
	case "/materials":
		material.Handle(chat)
		break
	case "/links":
		link.Handle(chat)
		break
	default:
		log.Printf("message command not matched")
		reply := helpers.GetReply(helpers.ErrorInvalidCommandRpl)
		tg.SendText(bot, chat.Id, reply)
	}
}

func HandleMessage(
	bot *tgbotapi.BotAPI,
	join *joinfeat.Handler, reminder *reminderfeat.Handler,
	chat *core.Chat, message *tgbotapi.Message) {
	log.Printf("Try match message with one of state")
	switch chat.State {
	case core.Start:
		join.HandleEmail(chat, message.Text)
		reminder.HandleNewChat(chat)
		break
	default:
		log.Printf("message state not matched")
		reply := helpers.GetReply(helpers.ErrorInvalidCommandRpl)
		tg.SendText(bot, chat.Id, reply)
	}
}

func HandleCallback(
	bot *tgbotapi.BotAPI,
	material *materialfeat.Handler,
	chat *core.Chat, event *core.Event, data string) {
	log.Printf("Try match callback with one of event")
	switch event.Type {
	case materialfeat.GetFileEvent:
		material.HandleGetFileEvent(chat, data)
		break
	default:
		helpers.HandleUnknownErr(bot, chat.Id, errors.New("callback event not matched"))
	}
}
