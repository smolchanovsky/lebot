package reminderfeat

import (
	"errors"
	"github.com/guregu/dynamo"
	"google.golang.org/api/calendar/v3"
	"lebot/cmd/student-bot/core"
	"log"
	"time"
)

type Reminder struct {
	ChatId int64
}

type ChatCal struct {
	ChatId int64
	CalId  string
}

type Service struct {
	calSrv *calendar.Service
	db     *dynamo.DB
}

func NewService(calSrv *calendar.Service, db *dynamo.DB) *Service {
	return &Service{calSrv: calSrv, db: db}
}

const defaultTimeZone = "Europe/Moscow"

var ErrCalConflict = errors.New("more than one calendar found")
var ErrCalNotFound = errors.New("more than one calendar found")

func (base *Service) InitNewChat(chat *core.Chat) error {
	cal, err := base.calSrv.Calendars.Insert(&calendar.Calendar{
		Summary:  chat.UserName,
		TimeZone: defaultTimeZone}).Do()
	if err != nil {
		return err
	}

	table := base.db.Table("chatCals")
	err = table.Put(ChatCal{CalId: cal.Id, ChatId: chat.Id}).Run()
	if err != nil {
		return err
	}

	_, err = base.calSrv.Acl.Insert(cal.Id, &calendar.AclRule{
		Role: "owner",
		Scope: &calendar.AclRuleScope{
			Type:  "user",
			Value: chat.TeacherEmail,
		}}).Do()
	if err != nil {
		return err
	}

	return nil
}

func (base *Service) GetLessonsSoon() ([]*Reminder, error) {
	now := time.Now()
	minTime := now.Format(time.RFC3339)
	maxTime := now.Add(time.Hour).Format(time.RFC3339)

	calList, err := base.calSrv.CalendarList.List().Do()
	if err != nil {
		return nil, err
	}

	table := base.db.Table("chatCals")

	var reminders []*Reminder
	for _, cal := range calList.Items {
		var chatCals []*ChatCal
		err := table.Scan().Filter("CalId = ?", cal.Id).All(&chatCals)
		if err != nil {
			log.Print("calendar not found in db: ", err)
			continue
		}
		if len(chatCals) == 0 {
			return nil, ErrCalNotFound
		}
		if len(chatCals) > 1 {
			return nil, ErrCalConflict
		}

		events, err := base.calSrv.Events.List(cal.Id).TimeZone(defaultTimeZone).
			TimeMin(minTime).TimeMax(maxTime).SingleEvents(true).Do()
		if err != nil {
			log.Print("error while obtain events: ", err)
			continue
		}

		for range events.Items {
			reminders = append(reminders, &Reminder{ChatId: chatCals[0].ChatId})
		}
	}

	return reminders, nil
}
