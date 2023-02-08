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

func GetLinks(drive *drive.Service, chat *core.Chat) ([]Link, error) {
	linkFolderQuery := "'%s' in writers and name = 'links' and mimeType = 'application/vnd.google-apps.folder'"
	linkFolders, err := drive.Files.
		List().
		Q(fmt.Sprintf(linkFolderQuery, chat.TeacherEmail)).
		PageSize(1).
		Do()
	if err != nil {
		return nil, err
	}

	linkFileQuery := "'%s' in parents"
	linkFiles, err := drive.Files.List().PageSize(10).
		Q(fmt.Sprintf(linkFileQuery, linkFolders.Files[0].Id)). // smolchanovsky@gmail.com
		Fields("nextPageToken, files(id, name)").
		Do()
	if err != nil {
		return nil, err
	}

	linkFile, err := drive.Files.Get(linkFiles.Files[0].Id).Download()
	if err != nil {
		return nil, err
	}

	defer linkFile.Body.Close()
	body, err := io.ReadAll(linkFile.Body)
	if err != nil {
		return nil, err
	}

	var links []Link
	yaml.Unmarshal(body, &links)
	return links, err
}
