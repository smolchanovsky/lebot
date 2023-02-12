package reply

import (
	"errors"
	"gopkg.in/yaml.v3"
	"os"
	"path"
	"path/filepath"
	"runtime"
)

type replica struct {
	Id   string
	Text string
}

var ErrReplicaNotFound = errors.New("reply not found")

func LoadReplica(configPath string, id string) (*string, error) {
	_, basePath, _, _ := runtime.Caller(0)
	baseDir := filepath.Join(filepath.Dir(basePath), "../../")

	fullPath := path.Join(baseDir, configPath)
	data, err := os.ReadFile(fullPath)
	if err != nil {
		return nil, err
	}

	var messages []*replica
	err = yaml.Unmarshal(data, &messages)
	if err != nil {
		return nil, err
	}

	for _, message := range messages {
		if message.Id == id {
			return &message.Text, nil
		}
	}

	return nil, ErrReplicaNotFound
}
