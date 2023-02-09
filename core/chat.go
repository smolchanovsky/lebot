package core

import (
	"errors"
	"github.com/guregu/dynamo"
)

var ErrChatConflict = errors.New("more than one chat found")

func GetChat(db *dynamo.DB, id int64) (*Chat, error) {
	table := db.Table("chats")

	var chats []*Chat
	err := table.Get("Id", id).All(&chats)
	if err != nil {
		return nil, err
	}

	if len(chats) == 0 {
		return nil, nil
	}

	if len(chats) > 1 {
		return nil, ErrChatConflict
	}

	return chats[0], err
}
