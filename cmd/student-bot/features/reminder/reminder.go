package reminder

import (
	"github.com/guregu/dynamo"
	"google.golang.org/api/calendar/v3"
	"lebot/cmd/student-bot/core"
	"log"
	"time"
)

type Reminder struct {
	ChatId int64
	Type   string
}

type ChatCal struct {
	ChatId int64
	CalId  string
}

const defaultTimeZone = "Europe/Moscow"

func Init(srv *calendar.Service, db *dynamo.DB, chat *core.Chat) error {
	cal, err := srv.Calendars.Insert(&calendar.Calendar{
		Summary:  chat.UserName,
		TimeZone: defaultTimeZone}).Do()
	if err != nil {
		return err
	}

	table := db.Table("chatCals")
	err = table.Put(ChatCal{CalId: cal.Id, ChatId: chat.Id}).Run()
	if err != nil {
		return err
	}

	_, err = srv.Acl.Insert(cal.Id, &calendar.AclRule{
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

func GetStartingLessons(srv *calendar.Service, db *dynamo.DB) ([]*Reminder, error) {
	now := time.Now()
	minTime := now.Format(time.RFC3339)
	maxTime := now.Add(time.Hour).Format(time.RFC3339)

	calList, err := srv.CalendarList.List().Do()
	if err != nil {
		return nil, err
	}

	table := db.Table("chatCals")

	reminders := []*Reminder{}
	for _, cal := range calList.Items {
		var chatCal *ChatCal
		err := table.Get("CalId", cal.Id).One(&chatCal)
		if err != nil {
			log.Print("calendar not found in db", err)
			continue
		}

		events, err := srv.Events.List(cal.Id).TimeZone(defaultTimeZone).
			TimeMin(minTime).TimeMax(maxTime).Do()
		if err != nil {
			log.Print("error while obtain events", err)
			continue
		}

		for range events.Items {
			reminders = append(reminders, &Reminder{ChatId: chatCal.ChatId, Type: "lessonSoon"})
		}
	}

	return reminders, nil
}
