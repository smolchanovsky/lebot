package joinfeat

import (
	"errors"
	"github.com/guregu/dynamo"
	"lebot/cmd/student-bot/core"
	"net/mail"
	"strings"
)

type Service struct {
	db *dynamo.DB
}

func NewService(db *dynamo.DB) *Service {
	return &Service{db: db}
}

func (base *Service) SaveChat(chatId int64, userName string) (*core.Chat, error) {
	table := base.db.Table("chats")

	chat := core.Chat{
		Id:           chatId,
		TeacherEmail: "",
		UserName:     userName,
		State:        core.Start,
	}

	err := table.Put(chat).Run()
	if err != nil {
		return nil, err
	}

	return &chat, nil
}

var ErrInvalidEmail = errors.New("invalid email")
var ErrEmailNotFound = errors.New("email not found")

func (base *Service) SaveTeacherEmail(chat *core.Chat, teacherEmail string) error {
	email := strings.ToLower(teacherEmail)

	_, err := mail.ParseAddress(email)
	if err != nil {
		return ErrInvalidEmail
	}

	teachers := base.db.Table("teachers")
	count, err := teachers.Get("Email", email).Count()
	if err != nil {
		return err
	}
	if count == 0 {
		return ErrEmailNotFound
	}

	chats := base.db.Table("chats")
	chat.TeacherEmail = strings.ToLower(email)
	chat.State = "ready"
	err = chats.Put(chat).Run()
	if err != nil {
		return err
	}

	return nil
}
