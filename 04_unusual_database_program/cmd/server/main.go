package main

import (
	"context"
	"flag"
	"fmt"
	"log/slog"
	"net"

	"github.com/ananthvk/protohackers-go/04_unusual_database_program/internal"
)

const maxPacketSize = 1000
const appVersion = "1.0.1"

func main() {
	portPtr := flag.Uint("port", 8000, "specify the port on which to listen")
	hostPtr := flag.String("host", "0.0.0.0", "specify the bind address")
	flag.Parse()
	address := fmt.Sprintf("%s:%d", *hostPtr, *portPtr)

	ctx := context.Background()
	listenerConfig := net.ListenConfig{}
	listener, err := listenerConfig.ListenPacket(ctx, "udp", address)
	if err != nil {
		slog.Error("listen failed", "error", err)
		return
	}
	store := internal.NewKVStore(appVersion)
	slog.Info("server listening", "address", listener.LocalAddr().String())
	defer listener.Close()
	buffer := make([]byte, maxPacketSize)
	for {
		n, fromAddr, err := listener.ReadFrom(buffer)
		result := store.ExecuteQuery(string(buffer[:n]))
		if result.HasValue {
			if _, err := listener.WriteTo([]byte(result.Value), fromAddr); err != nil {
				slog.Error("write error", "error", err)
			}
		}
		if err != nil {
			slog.Error("read error", "error", err)
		}
	}
}
