package core

const (
	GetMaterialEvent = 1
	GetLessonEvent   = 2
)

type ButtonEvent struct {
	Type  int    `json:"t"`
	Value string `json:"v"`
}
