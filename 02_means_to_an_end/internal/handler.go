package internal

import (
	"io"
	"log/slog"
	"net"
)

// Handle handles a single client connection. This should be run in a separate gorutine so that requests can be handled
// concurrently.
func Handle(connection net.Conn) {
	numRequests := int64(0)
	slog.Info("client connected", "remote_address", connection.RemoteAddr().String())
	defer func() {
		slog.Info("client disconnected", "address", connection.RemoteAddr().String(), "num_requests", numRequests)
	}()
	defer connection.Close()

	state := State{}
	buffer := make([]byte, messageSize)

	for {
		_, err := io.ReadFull(connection, buffer)
		if err != nil {
			return
		}
		numRequests++
		message, err := ParseMessage(buffer)
		if err != nil {
			return
		}
		if message.messageType == 'I' {
			state.Insert(message.field1, message.field2)
		} else {
			result := state.QueryAverage(message.field1, message.field2)
			if err := SendResponse(connection, Response{value: result}); err != nil {
				return
			}
		}
	}
}
