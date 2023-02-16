package materialfeat

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

var ErrMaterialsFolderNotFound = errors.New("materials folder not found")

func (base *Service) GetMaterialsRoot(chat *core.Chat) (*drive.File, error) {
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
		return nil, ErrMaterialsFolderNotFound
	}

	return materialsFolder.Files[0], err
}

func (base *Service) GetMaterials(folderId string) ([]*drive.File, error) {
	materialsQuery := "'%s' in parents"
	materialsFolder, err := base.diskSrv.Files.
		List().
		PageSize(10).
		Q(fmt.Sprintf(materialsQuery, folderId)).
		Do()
	if err != nil {
		return nil, err
	}

	return materialsFolder.Files, err
}

func (base *Service) GetMaterialMeta(id string) (*drive.File, error) {
	materialMeta, err := base.diskSrv.Files.Get(id).
		Fields("id", "name", "size", "webContentLink").Do()
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
