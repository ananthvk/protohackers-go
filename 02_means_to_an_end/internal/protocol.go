package internal

import (
	"bytes"
	"encoding/binary"
	"errors"
	"io"
)

const messageSize = 9 // bytes

type Message struct {
	messageType byte
	field1      int32
	field2      int32
}

type Response struct {
	value int32
}

func ParseMessage(b []byte) (Message, error) {
	if len(b) != 9 {
		return Message{}, errors.New("invalid payload length")
	}
	message := Message{}
	r := bytes.NewReader(b)
	if err := binary.Read(r, binary.BigEndian, &message.messageType); err != nil {
		return Message{}, err
	}
	if err := binary.Read(r, binary.BigEndian, &message.field1); err != nil {
		return Message{}, err
	}
	if err := binary.Read(r, binary.BigEndian, &message.field2); err != nil {
		return Message{}, err
	}
	if !(message.messageType == 'I' || message.messageType == 'Q') {
		return Message{}, errors.New("invalid message type")
	}
	return message, nil
}

func SendResponse(w io.Writer, response Response) error {
	return binary.Write(w, binary.BigEndian, response.value)
}
