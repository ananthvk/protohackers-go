package main

import (
	"context"
	"flag"
	"fmt"
	"log/slog"
	"net"

	"github.com/ananthvk/protohackers-go/06_speed_daemon/internal"
)

func main() {
	portPtr := flag.Uint("port", 8000, "specify the port on which to listen")
	hostPtr := flag.String("host", "0.0.0.0", "specify the bind address")
	flag.Parse()
	address := fmt.Sprintf("%s:%d", *hostPtr, *portPtr)

	ctx := context.Background()
	listenerConfig := net.ListenConfig{}
	listener, err := listenerConfig.Listen(ctx, "tcp", address)
	if err != nil {
		slog.Error("listen failed", "error", err)
		return
	}
	speedServer := internal.NewSpeedServer()
	slog.Info("server listening", "address", listener.Addr().String())
	defer listener.Close()
	for {
		conn, err := listener.Accept()
		if err != nil {
			slog.Warn("accept failed", "error", err)
			continue
		}
		go internal.Handle(speedServer, conn)
	}
}
