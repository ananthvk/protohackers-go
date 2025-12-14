package lrcp

import (
	"bytes"
	"errors"
)

// readNumeric reads a number, followed by a '/'. It returns the number of bytes processed (including the '/' character)
func readNumeric(b []byte) (int64, int, error) {
	var num int64
	var i int
	for i = 0; i < len(b); i++ {
		if b[i] < '0' || b[i] > '9' {
			break
		}
		digit := int64(b[i] - '0')
		if num > (2147483647-digit)/10 {
			return 0, -1, errors.New("number too large")
		}
		num = num*10 + digit
	}

	if i == 0 {
		return 0, -1, errors.New("number is empty")
	}

	if num >= 2147483648 {
		return 0, -1, errors.New("number too large")
	}

	// Read a '/' character at the end
	if i >= len(b) || b[i] != '/' {
		return 0, -1, errors.New("'/' not found following numeric field")
	}
	i++
	return num, i, nil
}

// parseString parses an unescaped string until it finds a '/' character.
// It returns slice if a string was found terminated by a '/' character
// It returns nil if the input slice is empty or there is no '/'
// It also returns a offset that represents the number of bytes processed (including terminating '/' character)
func parseString(b []byte) ([]byte, int) {
	for i := range b {
		if b[i] == '/' {
			return b[:i], i + 1
		}
	}
	return nil, -1
}

// Parses an escaped string (with \ characters), and consumes the terminating '/' character but does not add it to the returned slice
func parseEscapedString(b []byte) ([]byte, int, error) {
	idx := 0
	data := make([]byte, 0, min(1, len(b)))
	isEscaped := false
	for ; idx < len(b)-1; idx++ {
		if !isEscaped && (b[idx] == '/') {
			return data, -1, errors.New("expected '\\'")
		}
		if b[idx] == '\\' && !isEscaped {
			isEscaped = true
			continue
		}
		if isEscaped {
			if !(b[idx] == '\\' || b[idx] == '/') {
				return data, -1, errors.New("invalid escape sequence '\\' should be followed by '\\' or '/'")
			}
			isEscaped = false
		}
		data = append(data, b[idx])
	}
	return data, idx + 1, nil
}

func checkTrailingCharacters(m message, idx int, b []byte) (message, error) {
	if idx != len(b) {
		return message{}, errors.New("extra trailing characters")
	}
	return m, nil
}

func setMessageType(m *message, messageType []byte) error {
	if bytes.Equal(messageType, connectBytes) {
		m.kind = Connect
	} else if bytes.Equal(messageType, dataBytes) {
		m.kind = Data
	} else if bytes.Equal(messageType, ackBytes) {
		m.kind = Ack
	} else if bytes.Equal(messageType, closeBytes) {
		m.kind = Close
	} else {
		return errors.New("unknown message type")
	}
	return nil
}

func ParseMessage(b []byte) (message, error) {
	if len(b) >= 1000 {
		return message{}, errors.New("message too large")
	}
	if len(b) == 0 {
		return message{}, errors.New("empty message")
	}

	// Check that it's not empty
	var m message
	idx := 0

	// Check that it starts & ends with a '/'
	if b[idx] != '/' || b[len(b)-1] != '/' {
		return message{}, errors.New("message does not start and end with '/'")
	}
	idx++

	// Parse message type
	messageType, offset := parseString(b[idx:])
	if len(messageType) == 0 {
		return message{}, errors.New("message type not present")
	}
	idx += offset
	if err := setMessageType(&m, messageType); err != nil {
		return message{}, err
	}

	// Read session identifier
	session, offset, err := readNumeric(b[idx:])
	if err != nil {
		return message{}, err
	}
	m.sessionId = session
	idx += offset

	// If it's either a connect / close message, stop here
	if m.kind == Connect || m.kind == Close {
		return checkTrailingCharacters(m, idx, b)
	}

	// Read the next field, which is numeric and can either be POS or LENGTH
	pos_or_length, offset, err := readNumeric(b[idx:])
	if err != nil {
		return message{}, err
	}
	idx += offset
	if m.kind == Ack {
		m.length = pos_or_length
		return checkTrailingCharacters(m, idx, b)
	}

	m.pos = pos_or_length
	escapedString, offset, err := parseEscapedString(b[idx:])
	if err != nil {
		return message{}, err
	}
	idx += offset
	m.data = escapedString
	return checkTrailingCharacters(m, idx, b)
}
