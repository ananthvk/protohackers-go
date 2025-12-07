package tests

import (
	"bufio"
	"net"
	"strings"
	"testing"
	"time"

	"github.com/ananthvk/protohackers-go/03_budget_chat/internal"
)

// This tests the following:
// 1) The server sends a greeting message on connection to the client
// 2) The server reads a username from the client
// 3) The client "joins" the room
// 4) The server sends a presence notification (lists the users in the room, except the new user)

func TestJoinSingleClient(t *testing.T) {
	server, client := net.Pipe()
	chatServer := internal.NewChatServer()
	go func() {
		// Server
		internal.Handle(chatServer, server)
	}()

	reader := bufio.NewReader(client)
	writer := bufio.NewWriter(client)

	line, err := readLineWithTimeout(reader, time.Second*1)
	if err != nil {
		t.Errorf("unexpected error while receiving greeting message: %v", err)
		return
	}
	if !strings.Contains(line, "Welcome to budgetchat") {
		t.Errorf("want %q in greeting message, got %q", "Welcome to budgetchat", line)
		return
	}
	if err := writeStringWithTimeout(writer, "bob\n", time.Second*1); err != nil {
		t.Errorf("unexpected error while sending username: %v", err)
		return
	}
	line, err = readLineWithTimeout(reader, time.Second*1)
	if err != nil {
		t.Errorf("unexpected error while receiving presence notification: %v", err)
		return
	}
	if strings.Contains(line, "bob") {
		t.Errorf("presence notification contains the user who joined the room")
	}
}
