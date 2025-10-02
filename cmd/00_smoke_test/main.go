package main

import (
	"context"
	"flag"
	"fmt"
	"io"
	"log/slog"
	"net"
	"os"
	"os/signal"
	"sync"
	"sync/atomic"
	"syscall"
	"time"
)

const (
	clientShutdownTimeout = time.Second * 20
)

// TODO: Add client read timeouts (i.e. if the client hasn't sent anything in a while)

func main() {
	ctx, cancel := signal.NotifyContext(context.Background(), os.Interrupt, syscall.SIGTERM)
	defer cancel()

	portPtr := flag.Uint("port", 8000, "specify the port on which to listen")
	hostPtr := flag.String("host", "127.0.0.1", "specify the bind address")
	flag.Parse()

	var isClosed atomic.Bool
	isClosed.Store(false)

	address := fmt.Sprintf("%s:%d", *hostPtr, *portPtr)

	listenerConfig := net.ListenConfig{}
	listener, err := listenerConfig.Listen(ctx, "tcp", address)

	if err != nil {
		slog.Error("listen failed", "address", address, "error", err.Error())
		os.Exit(1)
	}
	defer listener.Close()

	slog.Info("server started listening", "address", address)

	var wg sync.WaitGroup

	wg.Add(1)
	go func() {
		for {
			conn, err := listener.Accept()
			if err != nil {
				if isClosed.Load() {
					slog.Info("stopped accepting new connections")
				} else {
					slog.Error("accept failed", "error", err)
					cancel()
				}
				break
			}
			wg.Add(1)
			go handleConnection(conn, &wg)
		}
		slog.Info("stopped accept loop")
		wg.Done()
	}()

	<-ctx.Done()
	isClosed.Store(true)
	// Close the listener so that we do not accept new connections
	listener.Close()

	// The next time we get an interrupt, it'll not be handled and the application closes
	cancel()

	slog.Info("shutting down, waiting for active connections to close", "time_left", clientShutdownTimeout)

	finished := make(chan struct{})

	go func() {
		wg.Wait()
		finished <- struct{}{}
	}()

	select {
	case <-finished:
		slog.Info("finished shutting down")
	case <-time.After(clientShutdownTimeout):
		slog.Warn("shutdown timeout exceeded, forcing exit")
	}
}

// handleConnection handles a single TCP connection to the server. It echoes back whatever is sent
// by the client
func handleConnection(conn net.Conn, wg *sync.WaitGroup) {
	defer wg.Done()
	addr := conn.RemoteAddr()
	slog.Info("accepted connection", "address", addr)
	n, err := io.Copy(conn, conn)
	if err != nil {
		slog.Error("echo failed", "address", addr, "error", err.Error())
	}
	err = conn.Close()
	if err != nil {
		slog.Error("client close failed", "address", addr, "error", err)
	}
	slog.Info("closed connection", "address", addr, "bytes_transferred", n)
}
