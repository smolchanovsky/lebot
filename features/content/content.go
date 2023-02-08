package content

import (
	"fmt"
	"google.golang.org/api/drive/v3"
	"io"
	"lebot/core"
)

const GetFileEvent = "getFileEvent"

type FileEvent struct {
	Type   string
	FileId string
}

func GetFileMeta(drive *drive.Service, id string) (*drive.File, error) {
	fileMeta, err := drive.Files.Get(id).Fields("id", "name", "size", "webContentLink").Do()
	return fileMeta, err
}

func GetFileContent(drive *drive.Service, id string) ([]byte, error) {
	response, err := drive.Files.Get(id).Download()
	if err != nil {
		return nil, err
	}

	defer response.Body.Close()
	content, err := io.ReadAll(response.Body)
	return content, err
}

func GetFiles(drive *drive.Service, chat *core.Chat) ([]*drive.File, error) {
	filesFolderQuery := "'%s' in writers and name = 'files' and mimeType = 'application/vnd.google-apps.folder'"
	filesFolder, err := drive.Files.
		List().
		PageSize(10).
		Q(fmt.Sprintf(filesFolderQuery, chat.TeacherEmail)).
		Do()
	if err != nil {
		return nil, err
	}

	filesQuery := "'%s' in parents"
	fileList, err := drive.Files.
		List().
		PageSize(10).
		Q(fmt.Sprintf(filesQuery, filesFolder.Files[0].Id)). // smolchanovsky@gmail.com
		Do()
	if err != nil {
		return nil, err
	}

	return fileList.Files, err
}
