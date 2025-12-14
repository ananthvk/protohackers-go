package lrcp

import (
	"fmt"
)

func escapeString(b []byte) []byte {
	buffer := make([]byte, 0, len(b))
	for _, byt := range b {
		if byt == '/' || byt == '\\' {
			// Escape the byte
			buffer = append(buffer, '\\')
		}
		buffer = append(buffer, byt)
	}
	return buffer
}

func SerializeMessage(m message) []byte {
	result := fmt.Appendf(nil, "/%s/%d/", m.kind, m.sessionId)

	switch m.kind {
	case Connect, Close:
		// Already complete
	case Ack:
		result = fmt.Appendf(nil, "/%s/%d/%d/", m.kind, m.sessionId, m.length)
	case Data:
		escapedData := escapeString(m.data)
		result = fmt.Appendf(nil, "/%s/%d/%d/", m.kind, m.sessionId, m.pos)
		result = append(result, escapedData...)
		result = append(result, '/')
	default:
		panic("Invalid message type")
	}
	return result
}
