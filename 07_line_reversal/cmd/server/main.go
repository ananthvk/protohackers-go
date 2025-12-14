package main

import (
	"context"
	"flag"
	"fmt"
	"log/slog"

	"github.com/ananthvk/protohackers-go/07_line_reversal/internal"
	"github.com/ananthvk/protohackers-go/07_line_reversal/internal/lrcp"
)

func main() {
	portPtr := flag.Uint("port", 8000, "specify the port on which to listen")
	hostPtr := flag.String("host", "0.0.0.0", "specify the bind address")
	flag.Parse()
	address := fmt.Sprintf("%s:%d", *hostPtr, *portPtr)

	ctx := context.Background()
	listenerConfig := lrcp.ListenConfig{}
	listener, err := listenerConfig.Listen(ctx, "lrcp", address)
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
		go internal.Handle(conn)
	}
}
