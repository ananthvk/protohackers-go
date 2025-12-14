package internal

import (
	"bufio"
	"log/slog"
	"net"
	"strings"
)

func reverseString(s string) string {
	r := []rune(s)
	for i, j := 0, len(r)-1; i < j; i, j = i+1, j-1 {
		r[i], r[j] = r[j], r[i]
	}
	return string(r)
}

// Handle handles a single client connection. This should be run in a separate gorutine so that requests can be handled
// concurrently.
func Handle(connection net.Conn) {
	slog.Info("client connected", "remote_address", connection.RemoteAddr().String())
	defer func() {
		slog.Info("client disconnected", "address", connection.RemoteAddr().String())
	}()
	defer connection.Close()

	r := bufio.NewReader(connection)
	w := bufio.NewWriter(connection)

	for {
		line, err := r.ReadString('\n')
		if err != nil {
			slog.Error("read error", "error", err)
			return
		}
		slog.Info("got line", "line", line)

		line = strings.TrimSuffix(line, "\n")

		line = reverseString(line) + "\n"
		slog.Info("reversed", "line", line)
		_, err = w.WriteString(line)
		if err != nil {
			slog.Error("write error", "error", err)
			return
		}
		err = w.Flush()
		if err != nil {
			slog.Error("flush error", "error", err)
			return
		}
	}
}
