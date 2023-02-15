package materialfeat

import (
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

func (base *Service) GetMaterialMeta(id string) (*drive.File, error) {
	materialMeta, err := base.diskSrv.Files.Get(id).Fields("id", "name", "size", "webContentLink").Do()
	return materialMeta, err
}

func (base *Service) GetMaterialContent(id string) ([]byte, error) {
	response, err := base.diskSrv.Files.Get(id).Download()
	if err != nil {
		return nil, err
	}

	defer response.Body.Close()
	content, err := io.ReadAll(response.Body)
	return content, err
}

func (base *Service) GetMaterials(chat *core.Chat) ([]*drive.File, error) {
	materialsFolderQuery := "'%s' in writers and name = 'materials' and mimeType = 'application/vnd.google-apps.folder'"
	materialsFolder, err := base.diskSrv.Files.
		List().
		PageSize(10).
		Q(fmt.Sprintf(materialsFolderQuery, chat.TeacherEmail)).
		Do()
	if err != nil {
		return nil, err
	}

	if len(materialsFolder.Files) == 0 {
		return []*drive.File{}, nil
	}

	materialsQuery := "'%s' in parents"
	materialList, err := base.diskSrv.Files.
		List().
		PageSize(10).
		Q(fmt.Sprintf(materialsQuery, materialsFolder.Files[0].Id)). // smolchanovsky@gmail.com
		Do()
	if err != nil {
		return nil, err
	}

	return materialList.Files, err
}
