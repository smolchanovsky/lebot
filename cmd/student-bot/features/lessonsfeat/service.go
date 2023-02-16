package lessonsfeat

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

var ErrLessonsFolderNotFound = errors.New("lesson folder not found")

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

func (base *Service) GetLessons(chat *core.Chat) ([]*drive.File, error) {
	lessonsFolder, err := base.getLessonsFolder(chat)
	if err != nil {
		return nil, err
	}

	studentFolder, err := base.getStudentFolder(chat, lessonsFolder.Id)
	if err != nil {
		return nil, err
	}

	lessonFilesQuery := "'%s' in parents"
	lessonFiles, err := base.diskSrv.Files.
		List().
		PageSize(10).
		Q(fmt.Sprintf(lessonFilesQuery, studentFolder.Id)).
		Do()
	if err != nil {
		return nil, err
	}

	return lessonFiles.Files, nil
}

func (base *Service) GetLessonContent(id string) ([]byte, error) {
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

func (base *Service) getStudentFolder(chat *core.Chat, lessonFolderId string) (*drive.File, error) {
	studentFolderQuery := "'%s' in writers and '%s' in parents and name = '%s' and mimeType = 'application/vnd.google-apps.folder'"
	studentFolder, err := base.diskSrv.Files.
		List().
		PageSize(10).
		Q(fmt.Sprintf(studentFolderQuery, chat.TeacherEmail, lessonFolderId, chat.UserName)).
		Do()
	if err != nil {
		return nil, err
	}
	if len(studentFolder.Files) == 0 {
		return nil, ErrLessonsFolderNotFound
	}

	return studentFolder.Files[0], nil
}
