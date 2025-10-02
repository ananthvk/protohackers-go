package tcp

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

type ClientHandler func(net.Conn) error

const (
	clientShutdownTimeout = time.Second * 20
)

// TODO: Add client read timeouts (i.e. if the client hasn't sent anything in a while)

type Server struct {
	Address       string
	listener      net.Listener
	portPtr       *uint
	hostPtr       *string
	isClosed      atomic.Bool
	ctx           context.Context
	wg            sync.WaitGroup
	clientHandler ClientHandler
}

// AddToFlags registers command-line flags for configuring the server's network settings.
// It adds a "port" flag to specify the listening port and a "host" flag to specify the address to listen on.
func (s *Server) AddToFlags() {
	s.portPtr = flag.Uint("port", 8000, "specify the port on which to listen")
	s.hostPtr = flag.String("host", "127.0.0.1", "specify the bind address")
}

// LoadFromFlags should be called after flag.Parse() is called in the calling function.
func (s *Server) LoadFromFlags() {
	s.Address = fmt.Sprintf("%s:%d", *s.hostPtr, *s.portPtr)
}

func (s *Server) Closed() bool {
	return s.isClosed.Load()
}

func NewServer(ctx context.Context) *Server {
	server := &Server{ctx: ctx}
	server.isClosed.Store(false)
	server.clientHandler = exampleEchoHandler
	return server
}

func (s *Server) ListenAndServe() {
	ctx, cancel := signal.NotifyContext(s.ctx, os.Interrupt, syscall.SIGTERM)
	defer cancel()

	listenerConfig := net.ListenConfig{}
	listener, err := listenerConfig.Listen(ctx, "tcp", s.Address)
	s.listener = listener
	if err != nil {
		slog.Error("listen failed", "address", s.Address, "error", err.Error())
		return
	}
	defer listener.Close()
	slog.Info("server started listening", "address", s.Address)

	s.wg.Add(1)
	go func() {
		for {
			conn, err := listener.Accept()
			if err != nil {
				if s.isClosed.Load() {
					slog.Info("stopped accepting new connections")
				} else {
					slog.Error("accept failed", "error", err)
					cancel()
				}
				break
			}
			s.wg.Add(1)
			go handleConnection(s, conn)
		}
		slog.Info("stopped accept loop")
		s.wg.Done()
	}()

	<-ctx.Done()
	s.isClosed.Store(true)
	// Close the listener so that we do not accept new connections
	listener.Close()

	// The next time we get an interrupt, it'll not be handled and the application closes
	cancel()

	slog.Info("shutting down, waiting for active connections to close", "time_left", clientShutdownTimeout)

	finished := make(chan struct{})

	go func() {
		s.wg.Wait()
		finished <- struct{}{}
	}()

	select {
	case <-finished:
		slog.Info("finished shutting down")
	case <-time.After(clientShutdownTimeout):
		slog.Warn("shutdown timeout exceeded, forcing exit")
	}
}

func (s *Server) SetClientHandler(handler ClientHandler) {
	s.clientHandler = handler
}

// handleConnection handles a single TCP connection. It calls the client handler to process the connection. Note: This function
// must be in a new goroutine otherwise it'll block.
func handleConnection(s *Server, conn net.Conn) {
	defer s.wg.Done()
	addr := conn.RemoteAddr()
	slog.Info("accepted connection", "address", addr)

	// Process the request
	startTime := time.Now()
	err := s.clientHandler(conn)
	duration := time.Since(startTime)

	if err != nil {
		slog.Error("client handler failed", "address", addr, "error", err.Error())
	}

	err = conn.Close()
	if err != nil {
		slog.Error("client close failed", "address", addr, "error", err, "session_duration", duration)
		return
	}
	slog.Info("closed connection", "address", addr, "session_duration", duration)
}

func exampleEchoHandler(conn net.Conn) error {
	n, err := io.Copy(conn, conn)
	if err != nil {
		return err
	}
	slog.Info("finished echo", "address", conn.RemoteAddr(), "bytes_transferred", n)
	return nil
}
