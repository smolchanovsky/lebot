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
	EventId   string
	ChatId    int64
	CreatedAt time.Time
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

	chatCalTable := base.db.Table("chatCals")
	reminderTable := base.db.Table("reminders")

	var reminders []*Reminder
	for _, cal := range calList.Items {
		var chatCals []*ChatCal
		err := chatCalTable.Scan().Filter("CalId = ?", cal.Id).All(&chatCals)
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

		for _, event := range events.Items {
			var dbReminders []*Reminder
			err = reminderTable.Get("EventId", event.Id).All(&dbReminders)
			if err != nil {
				log.Print("error while obtain reminder from db: ", err)
				continue
			}

			if len(dbReminders) > 0 {
				log.Print("skip already sent reminder")
				continue
			}

			newReminder := Reminder{
				EventId:   event.Id,
				ChatId:    chatCals[0].ChatId,
				CreatedAt: time.Now(),
			}
			err := reminderTable.Put(&newReminder).Run()
			if err != nil {
				log.Print("error while save reminder to db: ", err)
				continue
			}
			reminders = append(reminders, &newReminder)
		}
	}

	return reminders, nil
}
