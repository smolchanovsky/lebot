package reminderfeat

import (
	"errors"
	"fmt"
	"github.com/guregu/dynamo"
	"google.golang.org/api/calendar/v3"
	"lebot/cmd/student-bot/core"
	"log"
	"time"
)

type Reminder struct {
	Id        string
	EventId   string
	ChatId    int64
	CreatedAt time.Time
	Type      string
	Url       *string
}

const (
	LessonSoonType  = "LessonSoonType"
	LessonStartType = "LessonStartType"
)

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
	return base.getLessons(LessonSoonType, func(event *calendar.Event, now time.Time) (bool, error) {
		eventStart, err := time.Parse(time.RFC3339, event.Start.DateTime)
		if err != nil {
			log.Print("could not parse start date of event: ", err)
			return false, err
		}

		if eventStart.After(now) && eventStart.Sub(now) < time.Hour {
			return true, nil
		}

		return false, nil
	})
}

func (base *Service) GetLessonsStart() ([]*Reminder, error) {
	return base.getLessons(LessonStartType, func(event *calendar.Event, now time.Time) (bool, error) {
		eventStart, err := time.Parse(time.RFC3339, event.Start.DateTime)
		if err != nil {
			log.Print("could not parse start date of event: ", err)
			return false, err
		}

		if eventStart.After(now) && eventStart.Sub(now) < time.Minute*15 {
			return true, nil
		}

		return false, nil
	})
}

func (base *Service) getLessons(reminderType string, condition func(event *calendar.Event, now time.Time) (bool, error)) ([]*Reminder, error) {
	now := time.Now()
	minTime := now.Add(-time.Hour).Format(time.RFC3339)
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
			err = reminderTable.Get("Id", GetReminderId(event, reminderType)).All(&dbReminders)
			if err != nil {
				log.Print("error while obtain reminder from db: ", err)
				continue
			}

			if len(dbReminders) > 0 {
				log.Print("skip already sent reminder")
				continue
			}
			chatId := chatCals[0].ChatId

			shouldSend, err := condition(event, now)
			if err != nil {
				log.Print("error while invoke condition: ", err)
				continue
			}
			if !shouldSend {
				log.Print("skip event by condition")
				continue
			}

			var url *string = nil
			if len(event.ConferenceData.EntryPoints) > 0 {
				url = &event.ConferenceData.EntryPoints[0].Uri
			}

			newReminder := Reminder{
				Id:        fmt.Sprintf("%s_%s", event.Id, reminderType),
				EventId:   event.Id,
				ChatId:    chatId,
				CreatedAt: time.Now(),
				Type:      reminderType,
				Url:       url,
			}
			err = reminderTable.Put(&newReminder).Run()
			if err != nil {
				log.Print("error while save reminder to db: ", err)
				continue
			}
			reminders = append(reminders, &newReminder)
			log.Print("reminder will be sent")
		}
	}

	return reminders, nil
}

func GetReminderId(event *calendar.Event, reminderType string) string {
	return fmt.Sprintf("%s_%s", event.Id, reminderType)
}
