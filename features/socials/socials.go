package socials

import (
	"fmt"
	"google.golang.org/api/drive/v3"
	"gopkg.in/yaml.v3"
	"io"
	"lebot/core"
)

type Link struct {
	Name string
	Url  string
}

func GetLinks(disk *drive.Service, chat *core.Chat) ([]*Link, error) {
	linkFolderQuery := "'%s' in writers and name = 'links' and mimeType = 'application/vnd.google-apps.folder'"
	linkFolders, err := disk.Files.
		List().
		Q(fmt.Sprintf(linkFolderQuery, chat.TeacherEmail)).
		PageSize(1).
		Do()
	if err != nil {
		return nil, err
	}
	if len(linkFolders.Files) == 0 {
		return []*Link{}, nil
	}

	linkFileQuery := "'%s' in parents"
	linkFiles, err := disk.Files.List().PageSize(10).
		Q(fmt.Sprintf(linkFileQuery, linkFolders.Files[0].Id)). // smolchanovsky@gmail.com
		Fields("nextPageToken, files(id, name)").
		Do()
	if err != nil {
		return nil, err
	}
	if len(linkFiles.Files) == 0 {
		return []*Link{}, nil
	}

	linkFile, err := disk.Files.Get(linkFiles.Files[0].Id).Download()
	if err != nil {
		return nil, err
	}

	defer linkFile.Body.Close()
	body, err := io.ReadAll(linkFile.Body)
	if err != nil {
		return nil, err
	}

	var links []*Link
	err = yaml.Unmarshal(body, &links)
	if err != nil {
		return nil, err
	}

	return links, err
}
