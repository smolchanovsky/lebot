package core

type Chat struct {
	Id           int64
	TeacherEmail string
	UserName     string
	State        string
}

const (
	Start = "start" // c2 == 2
)

type Replica struct {
	Id   string
	Text string
}

type Event struct {
	Type int `json:"t"`
}

type Update struct {
	Id     string
	ChatId int64
	Text   string
	Json   string
}
