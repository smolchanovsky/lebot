package main

import (
	"cloud.google.com/go/dialogflow/apiv2/dialogflowpb"
	"encoding/json"
	"errors"
	"fmt"
	tgbotapi "github.com/go-telegram-bot-api/telegram-bot-api/v5"
	"github.com/robfig/cron/v3"
	"lebot/cmd/student-bot/core"
	"lebot/cmd/student-bot/features/joinfeat"
	"lebot/cmd/student-bot/features/lessonsfeat"
	"lebot/cmd/student-bot/features/linkfeat"
	"lebot/cmd/student-bot/features/materialfeat"
	"lebot/cmd/student-bot/features/reminderfeat"
	"lebot/cmd/student-bot/features/schedulefeat"
	"lebot/cmd/student-bot/helpers"
	"lebot/internal/dynamodb"
	"lebot/internal/googlecalendar"
	"lebot/internal/googledialogflow"
	"lebot/internal/googledrive"
	"lebot/internal/tg"
	"log"
	"os"
	"strconv"
	"time"
)

func main() {
	logfile := fmt.Sprintf("/var/log/lebot/logs-%s", time.Now().Format("2006-01-02"))
	f, err := os.OpenFile(logfile, os.O_RDWR|os.O_CREATE|os.O_APPEND, 0666)
	if err != nil {
		log.Fatalf("error opening file: %v", err)
	}
	defer f.Close()
	log.SetOutput(f)

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

	dfClient, err := googledialogflow.NewClient()
	if err != nil {
		log.Fatal(err)
	}

	bot, err := tg.NewBotApi()
	if err != nil {
		log.Fatal(err)
	}

	msgSender := helpers.NewMsgSender(bot, db)

	joinSrv := joinfeat.NewService(db)
	joinHandler := joinfeat.NewHandler(joinSrv, msgSender)

	scheduleSrv := schedulefeat.NewService(calSrv, db)
	scheduleHandler := schedulefeat.NewHandler(scheduleSrv, msgSender)

	linkSrv := linkfeat.NewService(diskSrv)
	linkHandler := linkfeat.NewHandler(linkSrv, msgSender)

	materialSrv := materialfeat.NewService(diskSrv)
	materialHandler := materialfeat.NewHandler(materialSrv, msgSender)

	lessonSrv := lessonsfeat.NewService(diskSrv)
	lessonHandler := lessonsfeat.NewHandler(lessonSrv, msgSender)

	reminderSrv := reminderfeat.NewService(calSrv, db)
	reminderHandler := reminderfeat.NewHandler(reminderSrv, msgSender)

	scheduler := cron.New()
	scheduler.AddFunc("*/20 * * * *", func() {
		reminderHandler.HandleLessonsSoon()
	})
	scheduler.AddFunc("*/5 * * * *", func() {
		reminderHandler.HandleLessonsStart()
	})
	scheduler.Start()

	updateConfig := tgbotapi.NewUpdate(0)
	updateConfig.Timeout = 60

	log.Print("listening updates...")
	updates := bot.GetUpdatesChan(updateConfig)

	for update := range updates {
		log.Printf("new update '%d'", update.UpdateID)
		helpers.SaveUserUpdate(db, &update)

		if update.Message != nil {

			chatId := update.Message.Chat.ID
			text := update.Message.Text
			log.Printf("new reply in '%d' chat: %s", chatId, text)

			chatOrNil, err := core.GetChat(db, chatId)
			if err != nil {
				helpers.HandleUnknownErr(msgSender, chatId, err)
				continue
			}
			// Command: /start
			if chatOrNil == nil {
				joinHandler.HandleCommand(update.Message.Chat)
				continue
			}
			log.Printf("start processing '%d' chat with new message: %s", chatId, text)

			intent, err := googledialogflow.DetectIntentText(dfClient, "lebot-376821", strconv.FormatInt(chatId, 10), text)
			if err != nil {
				helpers.HandleUnknownErr(msgSender, chatOrNil.Id, err)
				continue
			}

			HandleIntent(msgSender, joinHandler, scheduleHandler, lessonHandler,
				materialHandler, linkHandler, reminderHandler, intent, chatOrNil)
		} else if update.CallbackQuery != nil {
			chatId := update.CallbackQuery.Message.Chat.ID
			data := update.CallbackQuery.Data
			log.Printf("new callback in '%d' chat: %s", chatId, data)

			chatOrNil, err := core.GetChat(db, chatId)
			if err != nil {
				helpers.HandleUnknownErr(msgSender, chatId, err)
				continue
			}
			log.Printf("start processing '%d' chat with new callback: %s", chatId, data)

			var event *core.Event
			err = json.Unmarshal([]byte(data), &event)
			if err != nil {
				helpers.HandleUnknownErr(msgSender, chatId, err)
				continue
			}

			HandleCallback(msgSender, materialHandler, lessonHandler, chatOrNil, event, data)
		}
	}
}

func HandleIntent(
	msgSender *helpers.MsgSender,
	join *joinfeat.Handler, scheduleHandler *schedulefeat.Handler, lessonHandler *lessonsfeat.Handler,
	material *materialfeat.Handler, link *linkfeat.Handler, reminder *reminderfeat.Handler,
	intent *dialogflowpb.QueryResult, chat *core.Chat) {
	log.Printf("try match message with one of commands")
	switch true {
	case chat.State == core.Start:
		join.HandleEmail(chat, intent.QueryText)
		lessonHandler.HandleNewChat(chat)
		reminder.HandleNewChat(chat)
		break
	case intent.QueryText == "/schedule" || intent.Intent.Name == core.ShowScheduleIntent:
		trySendReplyForIntent(msgSender, chat, intent)
		scheduleHandler.Handle(chat)
		break
	case intent.QueryText == "/lessons" || intent.Intent.Name == core.ShowLessonsIntent:
		trySendReplyForIntent(msgSender, chat, intent)
		lessonHandler.HandleCommand(chat)
		break
	case intent.QueryText == "/materials" || intent.Intent.Name == core.ShowMaterialsIntent:
		trySendReplyForIntent(msgSender, chat, intent)
		material.HandleCommand(chat)
		break
	case intent.QueryText == "/links" || intent.Intent.Name == core.ShowLinksIntent:
		trySendReplyForIntent(msgSender, chat, intent)
		link.HandleCommand(chat)
		break
	default:
		trySendReplyForIntent(msgSender, chat, intent)
	}
}

func HandleCallback(
	msgSender *helpers.MsgSender,
	material *materialfeat.Handler, lessonHandler *lessonsfeat.Handler,
	chat *core.Chat, event *core.Event, data string) {
	log.Printf("try match callback with one of event")
	switch event.Type {
	case core.MaterialEvent:
		material.HandleButtonEvent(chat, data)
		break
	case core.LessonEvent:
		lessonHandler.HandleButtonEvent(chat, data)
		break
	default:
		helpers.HandleUnknownErr(msgSender, chat.Id, errors.New("callback event not matched"))
	}
}

func trySendReplyForIntent(msgSender *helpers.MsgSender, chat *core.Chat, intent *dialogflowpb.QueryResult) {
	if len(intent.QueryText) > 0 && intent.QueryText[0] == '/' {
		return
	}

	reply := intent.GetFulfillmentText()
	if reply != "" {
		msgSender.SendText(chat.Id, reply)
	}
}
