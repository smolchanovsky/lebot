package materialfeat

import (
	"fmt"
	"google.golang.org/api/drive/v3"
	"io"
	"lebot/cmd/student-bot/core"
)

const GetFileEvent = 1

type FileEvent struct {
	Type   int
	FileId string
}

type Service struct {
	diskSrv *drive.Service
}

func NewService(diskSrv *drive.Service) *Service {
	return &Service{diskSrv: diskSrv}
}

func (base *Service) GetFileMeta(id string) (*drive.File, error) {
	fileMeta, err := base.diskSrv.Files.Get(id).Fields("id", "name", "size", "webContentLink").Do()
	return fileMeta, err
}

func (base *Service) GetFileContent(id string) ([]byte, error) {
	response, err := base.diskSrv.Files.Get(id).Download()
	if err != nil {
		return nil, err
	}

	defer response.Body.Close()
	content, err := io.ReadAll(response.Body)
	return content, err
}

func (base *Service) GetFiles(chat *core.Chat) ([]*drive.File, error) {
	filesFolderQuery := "'%s' in writers and name = 'files' and mimeType = 'application/vnd.google-apps.folder'"
	filesFolder, err := base.diskSrv.Files.
		List().
		PageSize(10).
		Q(fmt.Sprintf(filesFolderQuery, chat.TeacherEmail)).
		Do()
	if err != nil {
		return nil, err
	}

	if len(filesFolder.Files) == 0 {
		return []*drive.File{}, nil
	}

	filesQuery := "'%s' in parents"
	fileList, err := base.diskSrv.Files.
		List().
		PageSize(10).
		Q(fmt.Sprintf(filesQuery, filesFolder.Files[0].Id)). // smolchanovsky@gmail.com
		Do()
	if err != nil {
		return nil, err
	}

	return fileList.Files, err
}
