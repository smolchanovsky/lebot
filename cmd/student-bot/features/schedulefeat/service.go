package schedulefeat

import (
	"github.com/guregu/dynamo"
	"google.golang.org/api/calendar/v3"
	"lebot/cmd/student-bot/core"
	"log"
	"time"
)

type Lesson struct {
	start time.Time
	end   time.Time
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

func (base *Service) GetLessons(chat *core.Chat, count int64) ([]*Lesson, error) {
	now := time.Now()
	minTime := now.Format(time.RFC3339)

	table := base.db.Table("chatCals")
	var chatCal *ChatCal
	err := table.Get("ChatId", chat.Id).One(&chatCal)
	if err != nil {
		return nil, err
	}

	events, err := base.calSrv.Events.List(chatCal.CalId).TimeZone(defaultTimeZone).
		TimeMin(minTime).SingleEvents(true).MaxResults(count).Do()

	var lessons []*Lesson
	for _, event := range events.Items {
		start, err := time.Parse(time.RFC3339, event.Start.DateTime)
		if err != nil {
			log.Print("could not parse start date of event", err)
			continue
		}
		end, err := time.Parse(time.RFC3339, event.End.DateTime)
		if err != nil {
			log.Print("could not parse end date of event", err)
			continue
		}
		lesson := &Lesson{start: start, end: end}
		lessons = append(lessons, lesson)
	}

	return lessons, nil
}
