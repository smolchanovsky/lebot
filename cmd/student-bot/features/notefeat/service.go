package notefeat

import (
	"errors"
	"fmt"
	"google.golang.org/api/drive/v3"
	"io"
	"lebot/cmd/student-bot/core"
)

type Service struct {
	diskSrv *drive.Service
}

func NewService(diskSrv *drive.Service) *Service {
	return &Service{diskSrv: diskSrv}
}

var ErrLessonsFolderNotFound = errors.New("email not found")

func (base *Service) InitNewChat(chat *core.Chat) error {
	lessonsFolder, err := base.getLessonsFolder(chat)
	if err != nil {
		return err
	}

	_, err = base.diskSrv.Files.Create(&drive.File{
		Name:     chat.UserName,
		MimeType: "application/vnd.google-apps.folder",
		Parents:  []string{lessonsFolder.Id},
	}).Do()
	if err != nil {
		return err
	}

	return nil
}

func (base *Service) GetNotes(chat *core.Chat) ([]*drive.File, error) {
	lessonsFolder, err := base.getLessonsFolder(chat)
	if err != nil {
		return nil, err
	}

	userFolder, err := base.getUserFolder(chat, lessonsFolder.Id)
	if err != nil {
		return nil, err
	}

	notesQuery := "'%s' in parents"
	noteList, err := base.diskSrv.Files.
		List().
		PageSize(10).
		Q(fmt.Sprintf(notesQuery, userFolder.Id)).
		Do()
	if err != nil {
		return nil, err
	}

	return noteList.Files, nil
}

func (base *Service) GetNoteContent(id string) ([]byte, error) {
	response, err := base.diskSrv.Files.Export(id, "text/plain").Download()
	if err != nil {
		return nil, err
	}

	defer response.Body.Close()
	content, err := io.ReadAll(response.Body)
	return content, err
}

func (base *Service) getLessonsFolder(chat *core.Chat) (*drive.File, error) {
	lessonsFolderQuery := "'%s' in writers and name = 'lessons' and mimeType = 'application/vnd.google-apps.folder'"
	lessonsFolders, err := base.diskSrv.Files.
		List().
		PageSize(10).
		Q(fmt.Sprintf(lessonsFolderQuery, chat.TeacherEmail)).
		Do()
	if err != nil {
		return nil, err
	}
	if len(lessonsFolders.Files) == 0 {
		return nil, ErrLessonsFolderNotFound
	}

	return lessonsFolders.Files[0], nil
}

func (base *Service) getUserFolder(chat *core.Chat, lessonsFolderId string) (*drive.File, error) {
	studentFolderQuery := "'%s' in writers and '%s' in parents and name = '%s' and mimeType = 'application/vnd.google-apps.folder'"
	studentFolder, err := base.diskSrv.Files.
		List().
		PageSize(10).
		Q(fmt.Sprintf(studentFolderQuery, chat.TeacherEmail, lessonsFolderId, chat.UserName)).
		Do()
	if err != nil {
		return nil, err
	}
	if len(studentFolder.Files) == 0 {
		return nil, ErrLessonsFolderNotFound
	}

	return studentFolder.Files[0], nil
}
