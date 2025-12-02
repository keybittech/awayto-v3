package util

import (
	"errors"
	"strconv"
	"strings"

	"github.com/keybittech/awayto-v3/go/pkg/types"
)

var MAX_SOCKET_MESSAGE_LENGTH = 65535
var malformedSocketError = errors.New("malformed socket id")

func GetColonJoined(userSub, connId string) string {
	return userSub + ":" + connId
}

func SplitColonJoined(id string) (string, string, error) {
	if len(id) <= 1 {
		return "", "", malformedSocketError
	}

	colonIdx := strings.Index(id, ":")
	if colonIdx != -1 && colonIdx < len(id)-1 {
		return id[0:colonIdx], id[colonIdx+1:], nil
	} else {
		return "", "", malformedSocketError
	}
}

func GenerateMessage(padTo int, message *types.SocketMessage) []byte {
	finish := RunTimer()
	defer finish()
	storeStr := "f"
	if message.Store {
		storeStr = "t"
	}

	historicalStr := "f"
	if message.Historical {
		historicalStr = "t"
	}

	actionNumber := strconv.Itoa(int(message.Action))

	paddedMessage := PaddedLen(padTo, len(actionNumber)) + actionNumber +
		PaddedLen(padTo, 1) + storeStr +
		PaddedLen(padTo, 1) + historicalStr +
		PaddedLen(padTo, len(message.Timestamp)) + message.Timestamp +
		PaddedLen(padTo, len(message.Topic)) + message.Topic +
		PaddedLen(padTo, len(message.Sender)) + message.Sender +
		PaddedLen(padTo, len(message.Payload)) + message.Payload

	return []byte(paddedMessage)
}

func ParseMessage(padTo, cursor int, data []byte) (int, string, error) {
	lenEnd := cursor + padTo

	if len(data) < lenEnd {
		return 0, "", ErrCheck(errors.New("length index out of range"))
	}

	valLen, _ := strconv.Atoi(string(data[cursor:lenEnd]))
	valEnd := lenEnd + valLen

	if len(data) < valEnd {
		return 0, "", ErrCheck(errors.New("value index out of range"))
	}

	val := string(data[lenEnd:valEnd])

	return valEnd, val, nil
}
