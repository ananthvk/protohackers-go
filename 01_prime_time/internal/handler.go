package internal

import (
	"bufio"
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

	lineReader := bufio.NewReader(connection)
	for {
		line, err := lineReader.ReadBytes('\n')
		numRequests++
		if err != nil {
			slog.Info("client disconnected", "remote_address", connection.RemoteAddr().String())
			return
		}
		request, err := ParseRequest(line)
		if err != nil {
			connection.Write([]byte(err.Error() + "\n"))
			return
		}
		isPrime := IsPrime(request.Number)
		response := Response{Method: "isPrime", Prime: isPrime}
		if err := SendResponse(connection, response); err != nil {
			return
		}
	}
}
