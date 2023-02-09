package core

import (
	"errors"
	"gopkg.in/yaml.v3"
	"log"
	"os"
	"path"
	"path/filepath"
	"runtime"
)

type Message struct {
	Id   string
	Text string
}

var ErrMessageNotFound = errors.New("message not found")

func GetMessage(id string) string {
	message, err := GetMessageOrDefault(id)
	if err != nil {
		log.Print("message not found", err)
		return "Oops, unknown error 😬 Please, inform your tutor"
	}
	return *message
}

func GetMessageOrDefault(id string) (*string, error) {
	_, basePath, _, _ := runtime.Caller(0)
	baseDir := filepath.Join(filepath.Dir(basePath), "../")

	fullPath := path.Join(baseDir, "resources/messages.yml")
	data, err := os.ReadFile(fullPath)
	if err != nil {
		return nil, err
	}

	var messages []*Message
	err = yaml.Unmarshal(data, &messages)
	if err != nil {
		return nil, err
	}

	for _, message := range messages {
		if message.Id == id {
			return &message.Text, nil
		}
	}

	return nil, ErrMessageNotFound
}
