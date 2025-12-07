package internal

import (
	"bufio"
	"log/slog"
	"net"
	"strings"
)

const outgoingChannelSize = 10

// WriteLineAndFlush writes the given string along with a newline ('\n'), then flushes the stream.
func WriteLineAndFlush(writer *bufio.Writer, message string) error {
	if _, err := writer.WriteString(message + "\n"); err != nil {
		return err
	}
	if err := writer.Flush(); err != nil {
		return err
	}
	return nil
}

// Handle handles a single client connection. This should be run in a separate gorutine so that requests can be handled
// concurrently.
func Handle(chatServer *ChatServer, connection net.Conn) {
	isJoined := false
	slog.Info("client connected", "remote_address", connection.RemoteAddr().String())
	client := &ClientConnection{
		conn:     connection,
		reader:   bufio.NewReader(connection),
		writer:   bufio.NewWriter(connection),
		outgoing: make(chan string, outgoingChannelSize),
	}
	defer func() {
		if isJoined {
			chatServer.BroadcastExcept(client.getKey(), formatNotification(client.username, "left the room"))
		}
		chatServer.RemoveUser(connection.RemoteAddr().String())
		slog.Info("client disconnected", "address", connection.RemoteAddr().String())
	}()
	defer connection.Close()

	// Send a greeting message
	if err := WriteLineAndFlush(client.writer, formatGreeting()); err != nil {
		return
	}
	// Get the username
	line, err := client.reader.ReadString('\n')
	if err != nil {
		return
	}
	client.username = strings.TrimSuffix(line, "\n")
	if !validateName(client.username) {
		WriteLineAndFlush(client.writer, formatNotification("!system", "invalid username"))
		return
	}

	// Send presence notification to this client
	// Note: We do not need to filter the returned list since the current user is not yet added to the map
	if err := WriteLineAndFlush(client.writer, formatUserList(chatServer.GetUsers())); err != nil {
		return
	}

	// The current goroutine becomes the reader goroutine, a new goroutine is created to handles writes for this client
	go func() {
		for msg := range client.outgoing {
			if err := WriteLineAndFlush(client.writer, msg); err != nil {
				return
			}
		}
	}()

	// Add the user to the map since they have "joined"
	chatServer.AddUser(client)
	isJoined = true

	// Broadcast a join notification to connected users
	chatServer.BroadcastExcept(client.getKey(), formatNotification(client.username, "joined the room"))

	for {
		line, err := client.reader.ReadString('\n')
		if err != nil {
			return
		}
		line = strings.TrimSuffix(line, "\n")
		chatServer.BroadcastExcept(client.getKey(), formatBroadcast(client.username, line))
	}
}
