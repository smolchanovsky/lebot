package greeting

import (
	"awesomeProject/core"
	"errors"
	"github.com/guregu/dynamo"
	"net/mail"
	"strings"
)

func CreateChat(db *dynamo.DB, chatId int64) (*core.Chat, error) {
	table := db.Table("chats")

	chat := &core.Chat{Id: chatId, TeacherEmail: "", State: core.Start}

	err := table.Put(chat).Run()
	if err != nil {
		return nil, err
	}

	return chat, nil
}

func GetGreeting(chat *core.Chat) string {
	return "Enter your teacher gmail"
}

var ErrInvalidEmail = errors.New("invalid email")

func SaveTeacherEmail(db *dynamo.DB, chat *core.Chat, teacherEmail string) error {
	_, err := mail.ParseAddress(teacherEmail)
	if err != nil {
		return ErrInvalidEmail
	}

	table := db.Table("chats")
	chat.TeacherEmail = strings.ToLower(teacherEmail)
	chat.State = "ready"
	err = table.Put(chat).Run()
	if err != nil {
		return err
	}

	return nil
}
