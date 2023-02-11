package content

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

func GetFileMeta(disk *drive.Service, id string) (*drive.File, error) {
	fileMeta, err := disk.Files.Get(id).Fields("id", "name", "size", "webContentLink").Do()
	return fileMeta, err
}

func GetFileContent(disk *drive.Service, id string) ([]byte, error) {
	response, err := disk.Files.Get(id).Download()
	if err != nil {
		return nil, err
	}

	defer response.Body.Close()
	content, err := io.ReadAll(response.Body)
	return content, err
}

func GetFiles(disk *drive.Service, chat *core.Chat) ([]*drive.File, error) {
	filesFolderQuery := "'%s' in writers and name = 'files' and mimeType = 'application/vnd.google-apps.folder'"
	filesFolder, err := disk.Files.
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
	fileList, err := disk.Files.
		List().
		PageSize(10).
		Q(fmt.Sprintf(filesQuery, filesFolder.Files[0].Id)). // smolchanovsky@gmail.com
		Do()
	if err != nil {
		return nil, err
	}

	return fileList.Files, err
}
