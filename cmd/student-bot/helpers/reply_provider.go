package helpers

import (
	"lebot/internal/reply"
	"log"
)

const (
	JoinStartRpl         = "join.start"
	JoinInvalidEmailRpl  = "join.invalidEmail"
	JoinEmailNotFoundRpl = "join.emailNotFound"
	JoinFinishRpl        = "join.finish"

	ScheduleNoLessonsRpl = "schedule.noLessons"

	ContentListSummaryRpl = "content.listSummary"
	ContentEmptyListRpl   = "content.emptyList"

	LinkListSummaryRpl = "link.listSummary"
	LinkEmptyListRpl   = "link.emptyList"

	ReminderLessonSoonRpl  = "reminder.lessonSoon"
	ReminderLessonStartRpl = "reminder.lessonStart"

	ErrorInvalidCommandRpl = "error.invalidCommand"
	ErrorUnknownRpl        = "error.unknown"
)

func GetReply(name string) string {
	const configPath = "./cmd/student-bot/resources/replicas.yml"
	replica, err := reply.LoadReplica(configPath, name)
	if err == nil {
		return *replica
	}
	log.Printf("replica '%s' not loaded", name)

	replica, err = reply.LoadReplica(configPath, ErrorUnknownRpl)
	if err != nil {
		return *replica
	}
	log.Printf("replica '%s' not loaded", name)

	return "Sorry, unknown error"
}
