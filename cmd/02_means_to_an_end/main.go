package main

import (
	"bytes"
	"context"
	"encoding/binary"
	"errors"
	"flag"
	"fmt"
	"io"
	"log/slog"
	"net"
	"sync"

	"github.com/ananthvk/protohackers-go/internal/tcp"
	"github.com/google/uuid"
)

type Message struct {
	Type   byte
	Field1 int32
	Field2 int32
}

const messageSize = 9 // 9 bytes

type Stock struct {
	Timestamp int32
	Price     int32
}

type Stocks []Stock

var store map[uuid.UUID]Stocks
var mt sync.Mutex

// TODO: Brute force method to calculate mean, find if an efficient solution exists
func queryMean(clientId uuid.UUID, mintime, maxtime int32) int32 {
	mt.Lock()
	stocks, ok := store[clientId]
	mt.Unlock()
	if !ok {
		slog.Info("query failed since client does not exist", "clientId", clientId)
		return 0
	}
	total := int64(0)
	count := 0
	for _, stock := range stocks {
		if stock.Timestamp >= mintime && stock.Timestamp <= maxtime {
			total += int64(stock.Price)
			count++
		}
	}
	if count == 0 {
		return 0
	}
	return int32(total / int64(count))
}

func insertPrice(clientId uuid.UUID, timestamp, price int32) {
	mt.Lock()
	defer mt.Unlock()
	store[clientId] = append(store[clientId], Stock{Timestamp: timestamp, Price: price})
}

func deleteClient(clientId uuid.UUID) {
	mt.Lock()
	defer mt.Unlock()
	delete(store, clientId)
}

func createClient(clientId uuid.UUID) {
	mt.Lock()
	defer mt.Unlock()
	store[clientId] = make([]Stock, 0)
}

func handleClient(conn net.Conn) error {
	id := uuid.New()
	createClient(id)
	defer deleteClient(id)
	buffer := make([]byte, messageSize)
	for {
		_, err := io.ReadFull(conn, buffer)
		if err != nil {
			slog.Warn("unexpected message size", "address", conn.RemoteAddr())
			return err
		}
		var message Message
		err = binary.Read(bytes.NewReader(buffer), binary.BigEndian, &message)
		if err != nil {
			return errors.New("invalid message sent by client")
		}
		switch message.Type {
		case 'I':
			insertPrice(id, message.Field1, message.Field2)
		case 'Q':
			result := queryMean(id, message.Field1, message.Field2)
			err := binary.Write(conn, binary.BigEndian, result)
			if err != nil {
				return fmt.Errorf("error while sending result to client: %s", err.Error())
			}
		default:
			return fmt.Errorf("invalid query type %q", message.Type)
		}
	}
}

func main() {
	store = make(map[uuid.UUID]Stocks)
	server := tcp.NewServer(context.Background())
	server.AddToFlags()
	flag.Parse()
	server.LoadFromFlags()
	server.SetClientHandler(handleClient)
	server.ListenAndServe()
}
