package helpers

import (
	"log"
)

func HandleUnknownErr(msgSender *MsgSender, chatId int64, err error) {
	log.Print("unknown error: ", err)
	text := GetReply(ErrorUnknownRpl)
	msgSender.SendText(chatId, text)
}
