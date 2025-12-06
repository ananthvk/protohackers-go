package internal

import (
	"io"
	"log/slog"
	"net"
)

// Handle handles a single client connection. This should be run in a separate gorutine so that requests can be handled
// concurrently.
func Handle(connection net.Conn) {
	numBytes := int64(0)
	slog.Info("client connected", "remote_address", connection.RemoteAddr().String())
	defer func() {
		slog.Info("client disconnected", "address", connection.RemoteAddr().String(), "num_bytes", numBytes)
	}()
	defer connection.Close()
	numBytes, err := io.Copy(connection, connection)
	if err != nil {
		slog.Error("copy failed", "error", err)
		return
	}
}
