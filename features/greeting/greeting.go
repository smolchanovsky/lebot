package greeting

import (
	"errors"
	"github.com/guregu/dynamo"
	"lebot/core"
	"net/mail"
	"strings"
)

func CreateChat(db *dynamo.DB, chatId int64) (*core.Chat, error) {
	table := db.Table("chats")

	chat := core.Chat{Id: chatId, TeacherEmail: "", State: core.Start}

	err := table.Put(chat).Run()
	if err != nil {
		return nil, err
	}

	return &chat, nil
}

var ErrInvalidEmail = errors.New("invalid email")
var ErrEmailNotFound = errors.New("email not found")

func SaveTeacherEmail(db *dynamo.DB, chat *core.Chat, teacherEmail string) error {
	email := strings.ToLower(teacherEmail)

	_, err := mail.ParseAddress(email)
	if err != nil {
		return ErrInvalidEmail
	}

	teachers := db.Table("teachers")
	count, err := teachers.Get("Email", email).Count()
	if err != nil {
		return err
	}
	if count == 0 {
		return ErrEmailNotFound
	}

	chats := db.Table("chats")
	chat.TeacherEmail = strings.ToLower(email)
	chat.State = "ready"
	err = chats.Put(chat).Run()
	if err != nil {
		return err
	}

	return nil
}
