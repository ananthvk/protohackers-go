package main

import (
	"context"
	"flag"
	"fmt"
	"log/slog"
	"net"

	"github.com/ananthvk/protohackers-go/05_mob_in_the_middle/internal"
)

func main() {
	portPtr := flag.Uint("port", 8000, "specify the port on which to listen")
	hostPtr := flag.String("host", "0.0.0.0", "specify the bind address")
	upstreamPortPtr := flag.Uint("upstream-port", 16963, "specify the upstream host port")
	upstreamHostPtr := flag.String("upstream-host", "chat.protohackers.com", "specify the upstream host address")
	flag.Parse()
	address := fmt.Sprintf("%s:%d", *hostPtr, *portPtr)
	upstreamAddress := fmt.Sprintf("%s:%d", *upstreamHostPtr, *upstreamPortPtr)

	ctx := context.Background()
	listenerConfig := net.ListenConfig{}
	listener, err := listenerConfig.Listen(ctx, "tcp", address)
	if err != nil {
		slog.Error("listen failed", "error", err)
		return
	}
	slog.Info("server listening", "address", listener.Addr().String())
	defer listener.Close()
	for {
		conn, err := listener.Accept()
		if err != nil {
			slog.Warn("accept failed", "error", err)
			continue
		}
		go internal.Handle(upstreamAddress, conn)
	}
}
