package internal

import (
	"log/slog"
	"net"
)

const targetBogusCoin = "7YWHMfk9JZe0LM0g1ZauHuiSxhI"

func InterceptMessage(b []byte) []byte {
	return []byte(ReplaceBoguscoin(string(b), targetBogusCoin))
}

// Handle handles a single client connection. This should be run in a separate gorutine so that requests can be handled
// concurrently.
func Handle(upstreamAddress string, connection net.Conn) {
	slog.Info("client connected", "remote_address", connection.RemoteAddr().String())
	upstreamConn, err := net.Dial("tcp", upstreamAddress)
	if err != nil {
		slog.Error("proxy session creation failed", "error", err)
		return
	}
	session := NewProxySession(connection, upstreamConn)
	session.SetOnLineReceivedHandler(InterceptMessage)
	session.Start()
}
