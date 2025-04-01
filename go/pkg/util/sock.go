package util

import (
	"crypto/sha1"
	"encoding/base64"
	"encoding/binary"
	"errors"
	"fmt"
	"net"
	"strings"
)

func GetSocketId(userSub, connId string) string {
	return fmt.Sprintf("%s:%s", userSub, connId)
}

func SplitSocketId(id string) (string, string, error) {
	colonIdx := strings.Index(id, ":")
	if colonIdx == -1 {
		return "", "", ErrCheck(errors.New("could not parse socket id " + id))
	}
	return id[0:colonIdx], id[colonIdx:], nil
}

func ComputeWebSocketAcceptKey(clientKey string) string {
	const websocketGUID = "258EAFA5-E914-47DA-95CA-C5AB0DC85B11"
	h := sha1.New()
	h.Write([]byte(clientKey + websocketGUID))
	return base64.StdEncoding.EncodeToString(h.Sum(nil))
}

func ReadSocketConnectionMessage(conn net.Conn) ([]byte, error) {
	// Read header byte by byte to minimize blocking
	headerByte1 := make([]byte, 1)
	if _, err := conn.Read(headerByte1); err != nil {
		return nil, err
	}

	headerByte2 := make([]byte, 1)
	if _, err := conn.Read(headerByte2); err != nil {
		return nil, err
	}

	// Extract mask and payload length from second byte
	mask := headerByte2[0]&0x80 == 0x80
	payloadLen := int(headerByte2[0] & 0x7F)

	// Calculate total header size based on payload length indicator
	var extendedLenSize int
	if payloadLen == 126 {
		extendedLenSize = 2
	} else if payloadLen == 127 {
		extendedLenSize = 8
	}

	// Size for mask key if present
	maskSize := 0
	if mask {
		maskSize = 4
	}

	// Calculate total buffer size needed for the rest of the header
	headerExtraSize := extendedLenSize + maskSize

	// Read extended length and mask in one operation if needed
	headerExtra := make([]byte, headerExtraSize)
	if headerExtraSize > 0 {
		bytesRead := 0
		for bytesRead < headerExtraSize {
			n, err := conn.Read(headerExtra[bytesRead:])
			if err != nil {
				return nil, err
			}
			bytesRead += n
		}
	}

	// Parse extended payload length if needed
	actualPayloadLen := payloadLen
	if payloadLen == 126 {
		actualPayloadLen = int(binary.BigEndian.Uint16(headerExtra[:2]))
	} else if payloadLen == 127 {
		actualPayloadLen = int(binary.BigEndian.Uint64(headerExtra[:8]))
	}

	// Extract mask key if present
	var maskKey [4]byte
	if mask {
		copy(maskKey[:], headerExtra[extendedLenSize:extendedLenSize+4])
	}

	// Read payload directly
	data := make([]byte, actualPayloadLen)
	bytesRead := 0
	for bytesRead < actualPayloadLen {
		n, err := conn.Read(data[bytesRead:])
		if err != nil {
			return nil, err
		}
		bytesRead += n
	}

	// Apply mask if necessary
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
		return ErrCheck(err)
	}

	return nil
}
