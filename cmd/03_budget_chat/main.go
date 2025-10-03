package main

import (
	"bufio"
	"context"
	"flag"
	"fmt"
	"log/slog"
	"net"
	"strings"
	"sync"

	"github.com/ananthvk/protohackers-go/internal/tcp"
	"github.com/google/uuid"
)

type ChatServer struct {
	clients map[uuid.UUID]*ChatClient
	mt      sync.Mutex
}

type ChatClient struct {
	username string
	reader   *bufio.Reader
	writer   *bufio.Writer
	conn     net.Conn
	isJoined bool
	id       uuid.UUID
}

const welcomeMessage = "Hi there, welcome to chat app, To start chatting, please enter your name and press ENTER"

func NewChatServer() *ChatServer {
	return &ChatServer{
		clients: map[uuid.UUID]*ChatClient{},
	}
}

func isValidUsername(username string) error {
	numberOfCharacters := 0
	for _, char := range username {
		if (char >= 'A' && char <= 'Z') ||
			(char >= 'a' && char <= 'z') {
			numberOfCharacters++
		} else if char >= '0' && char <= '9' {
			// A number
		} else {
			return fmt.Errorf("invalid username: character %q in username", char)
		}
	}
	if numberOfCharacters == 0 {
		return fmt.Errorf("invalid username: username must contain atleast one character")
	}
	return nil
}

func (c *ChatServer) sendLine(client *ChatClient, line string, addNewline bool) error {
	if addNewline {
		line = line + "\n"
	}
	_, err := client.writer.WriteString(line)
	if err != nil {
		slog.Warn("failed to send message to client", "error", err)
		return err
	}
	err = client.writer.Flush()
	if err != nil {
		slog.Warn("failed to flush message", "error", err)
		return err
	}
	return nil
}

func (c *ChatServer) readLine(client *ChatClient) (string, error) {
	line, err := client.reader.ReadString('\n')
	if err != nil {
		return "", err
	}
	return line, nil
}

func (c *ChatServer) startClientJoinFlow(client *ChatClient) error {
	if err := c.sendLine(client, welcomeMessage, true); err != nil {
		return err
	}
	username, err := c.readLine(client)
	if err != nil {
		return err
	}
	username = strings.TrimSuffix(username, "\n")
	if err := isValidUsername(username); err != nil {
		err2 := c.sendLine(client, err.Error(), true)
		if err2 != nil {
			return err2
		}
		return err
	}
	client.username = username
	client.isJoined = true

	c.mt.Lock()
	id := uuid.New()
	client.id = id
	c.clients[id] = client
	c.mt.Unlock()

	slog.Info("client joined", "username", client.username, "id", id)
	return nil
}

func (c *ChatServer) startClientLeaveFlow(client *ChatClient) {
	if client.isJoined {
		c.broadcastMessage(fmt.Sprintf("%s left the room\n", client.username), true, client.id)
		c.mt.Lock()
		delete(c.clients, client.id)
		c.mt.Unlock()
		slog.Info("client left", "username", client.username, "id", client.id)
	}
}

func (c *ChatServer) broadcastMessage(message string, isSystemMessage bool, from uuid.UUID) {
	if isSystemMessage {
		message = fmt.Sprintf("* %s", message)
	} else {
		c.mt.Lock()
		client, ok := c.clients[from]
		c.mt.Unlock()
		if !ok {
			// The client does not exist
			slog.Error("broadcast failed since client does not exist", "id", from)
			// TOOD: Decide whether to continue the broadcast or not
			return
		}
		message = fmt.Sprintf("[%s] %s", client.username, message)
	}

	// TODO: Find a more efficient way to broadcast since we holding the lock for a long time
	c.mt.Lock()
	for _, client := range c.clients {
		if client.id == from {
			continue
		}
		c.sendLine(client, message, false)
	}
	c.mt.Unlock()
}

func (c *ChatServer) connectionHandler(conn net.Conn) error {
	reader := bufio.NewReader(conn)
	writer := bufio.NewWriter(conn)
	client := &ChatClient{reader: reader, conn: conn, writer: writer, isJoined: false}
	err := c.startClientJoinFlow(client)
	if err != nil {
		return err
	}
	c.mt.Lock()
	usernames := make([]string, 0, len(c.clients))
	for _, user := range c.clients {
		if user.id == client.id {
			continue
		}
		usernames = append(usernames, user.username)
	}
	c.mt.Unlock()
	users := strings.Join(usernames, " ")
	c.sendLine(client, fmt.Sprintf("* The room contains: %s", users), true)
	c.broadcastMessage(fmt.Sprintf("%s joined the room\n", client.username), true, client.id)
	defer c.startClientLeaveFlow(client)
	for {
		message, err := c.readLine(client)
		if err != nil {
			return err
		}
		c.broadcastMessage(message, false, client.id)
	}
}

func main() {
	server := tcp.NewServer(context.Background())
	server.AddToFlags()
	flag.Parse()
	server.LoadFromFlags()
	chatServer := NewChatServer()
	server.SetClientHandler(chatServer.connectionHandler)
	server.ListenAndServe()
}
