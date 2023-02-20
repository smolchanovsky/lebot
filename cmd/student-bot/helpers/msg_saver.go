package helpers

import (
	"encoding/json"
	"fmt"
	tgbotapi "github.com/go-telegram-bot-api/telegram-bot-api/v5"
	"github.com/google/uuid"
	"github.com/guregu/dynamo"
	"time"
)

const (
	MessageType  = "message"
	CallbackType = "callback"
	MsgReplyType = "messageReply"
	DocReplyType = "messageReply"
)

type message struct {
	Id        string
	Type      string
	CreatedAt time.Time
	ChatId    int64
	Text      string
	Json      string
}

func SaveMsgReply(db *dynamo.DB, msgConfig *tgbotapi.MessageConfig) error {
	table := db.Table("messages")
	msgJson, err := json.Marshal(msgConfig)
	if err != nil {
		return err
	} else {
		msg := &message{
			Id:        fmt.Sprintf("%d_%s", msgConfig.ChatID, uuid.New().String()),
			Type:      MsgReplyType,
			CreatedAt: time.Now(),
			ChatId:    msgConfig.ChatID,
			Text:      msgConfig.Text,
			Json:      string(msgJson),
		}
		err := table.Put(msg).Run()
		if err != nil {
			return err
		}
	}
	return nil
}

func SaveDocReply(db *dynamo.DB, docConfig *tgbotapi.DocumentConfig) error {
	table := db.Table("messages")
	msg := &message{
		Id:        fmt.Sprintf("%d_%s", docConfig.ChatID, uuid.New().String()),
		Type:      DocReplyType,
		CreatedAt: time.Now(),
		ChatId:    docConfig.ChatID,
		Text:      "",
		Json:      "",
	}
	err := table.Put(msg).Run()
	if err != nil {
		return err
	}
	return nil
}

func SaveUpdate(db *dynamo.DB, update *tgbotapi.Update) error {
	table := db.Table("messages")
	updateJson, err := json.Marshal(update)
	if err != nil {
		return err
	} else {
		var msg *message
		if update.Message != nil {
			msg = &message{
				Id:        fmt.Sprintf("%d_%d", update.Message.Chat.ID, update.Message.MessageID),
				Type:      MessageType,
				CreatedAt: time.Now(),
				ChatId:    update.Message.Chat.ID,
				Text:      update.Message.Text,
				Json:      string(updateJson),
			}
		} else {
			msg = &message{
				Id:        fmt.Sprintf("%d_%s", update.CallbackQuery.Message.Chat.ID, update.CallbackQuery.ID),
				Type:      CallbackType,
				CreatedAt: time.Now(),
				ChatId:    update.CallbackQuery.Message.Chat.ID,
				Text:      update.CallbackQuery.Data,
				Json:      string(updateJson),
			}
		}
		err := table.Put(msg).Run()
		if err != nil {
			return err
		}
	}
	return nil
}
