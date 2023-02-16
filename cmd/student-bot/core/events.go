package core

const (
	MaterialEvent = 1
	LessonEvent   = 2
)

const (
	GetFolderAction = 1
	GetFileAction   = 2
)

type ButtonEvent struct {
	Type   int    `json:"t"`
	Action int    `json:"a"`
	Value  string `json:"v"`
}
