package util

import (
	"crypto/sha1"
	"encoding/base64"
	"encoding/binary"
	"errors"
	"fmt"
	"io"
	"net"
	"strings"
)

func GetSocketId(userSub, connId string) string {
	return fmt.Sprintf("%s:%s", userSub, connId)
}

func SplitSocketId(id string) (string, string, error) {
	idParts := strings.Split(id, ":")
	if len(idParts) != 2 {
		errStr := fmt.Sprintf("id not in x:y format, received: %s", id)
		return "", "", ErrCheck(errors.New(errStr))
	}
	return idParts[0], idParts[1], nil
}

func ComputeWebSocketAcceptKey(clientKey string) string {
	const websocketGUID = "258EAFA5-E914-47DA-95CA-C5AB0DC85B11"
	h := sha1.New()
	h.Write([]byte(clientKey + websocketGUID))
	return base64.StdEncoding.EncodeToString(h.Sum(nil))
}

func ReadSocketConnectionMessage(conn net.Conn) ([]byte, error) {
	header := make([]byte, 2)
	_, err := io.ReadFull(conn, header)
	if err != nil {
		return nil, err
	}

	mask := header[1]&0x80 == 0x80
	payloadLen := int(header[1] & 0x7F)

	if payloadLen == 126 {
		lenBytes := make([]byte, 2)
		_, err := io.ReadFull(conn, lenBytes)
		if err != nil {
			return nil, err
		}
		payloadLen = int(binary.BigEndian.Uint16(lenBytes))
	} else if payloadLen == 127 {
		lenBytes := make([]byte, 8)
		_, err := io.ReadFull(conn, lenBytes)
		if err != nil {
			return nil, err
		}
		payloadLen = int(binary.BigEndian.Uint64(lenBytes))
	}

	var maskKey [4]byte
	if mask {
		_, err := io.ReadFull(conn, maskKey[:])
		if err != nil {
			return nil, err
		}
	}

	data := make([]byte, payloadLen)
	_, err = io.ReadFull(conn, data)
	if err != nil {
		return nil, err
	}

	if mask {
		for i := 0; i < len(data); i++ {
			data[i] ^= maskKey[i%4]
		}
	}

	return data, nil
}

func WriteSocketConnectionMessage(msg []byte, conn net.Conn) error {
	var frame []byte

	finAndOpcode := byte(0x81)
	frame = append(frame, finAndOpcode)

	length := len(msg)
	if length <= 125 {
		frame = append(frame, byte(length))
	} else if length <= 65535 {
		frame = append(frame, byte(126))
		frame = append(frame, byte(length>>8), byte(length&0xFF))
	} else {
		frame = append(frame, byte(127))
		for i := 7; i >= 0; i-- {
			frame = append(frame, byte((length>>(i*8))&0xFF))
		}
	}

	frame = append(frame, msg...)

	_, err := conn.Write(frame)
	if err != nil {
		return err
	}

	return nil
}
