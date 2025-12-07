package internal

import (
	"bufio"
	"net"
	"sync"
)

// ClientConnection is a connection from a single client. It has the net.Conn object, along with the buffered readers & writers for
// this particular connection
type ClientConnection struct {
	conn     net.Conn
	reader   *bufio.Reader
	writer   *bufio.Writer
	outgoing chan string
	username string
}

func (c *ClientConnection) getKey() string {
	return c.conn.RemoteAddr().String()
}

// ChatServer represents the global state of the chat application. It holds a mutex, and a map of clients
type ChatServer struct {
	mu sync.RWMutex
	// clients is a map of connection address to the connection object
	clients map[string]*ClientConnection
}

func NewChatServer() *ChatServer {
	return &ChatServer{
		clients: map[string]*ClientConnection{},
	}
}

func (c *ChatServer) AddUser(conn *ClientConnection) {
	c.mu.Lock()
	defer c.mu.Unlock()
	c.clients[conn.getKey()] = conn
}

func (c *ChatServer) RemoveUser(key string) {
	c.mu.Lock()
	defer c.mu.Unlock()
	client := c.clients[key]
	if client != nil {
		close(client.outgoing)
		client.conn.Close()
	}
	delete(c.clients, key)
}

func (c *ChatServer) GetUsers() []string {
	c.mu.RLock()
	defer c.mu.RUnlock()
	usernames := make([]string, 0, len(c.clients))
	for _, v := range c.clients {
		usernames = append(usernames, v.username)
	}
	return usernames
}

func (c *ChatServer) BroadcastExcept(excludeKey, message string) {
	// Copy the keys under RLock
	c.mu.RLock()
	recipients := make([]*ClientConnection, 0, len(c.clients))
	for k, client := range c.clients {
		if k != excludeKey {
			recipients = append(recipients, client)
		}
	}
	c.mu.RUnlock()

	// Send the messages to the outgoing channels of the connections
	for _, client := range recipients {
		select {
		case client.outgoing <- message:
		default:
			// Skip slow client
		}
	}
}
