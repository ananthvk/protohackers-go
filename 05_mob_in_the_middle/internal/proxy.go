package internal

import (
	"bufio"
	"log/slog"
	"net"
	"sync"
)

type ProxySession struct {
	// The client connection
	conn net.Conn

	// The upstream connection
	upstreamConn net.Conn

	clientRW   *bufio.ReadWriter
	upstreamRW *bufio.ReadWriter

	onLineReceivedHandler func([]byte) []byte
}

// NewProxySession creates a new proxy session. It creates  creates the buffered reader/writers.
// After creation of proxy session, normal Read/Write should not be used since they are buffered
// internally
func NewProxySession(conn net.Conn, upstreamConn net.Conn) *ProxySession {
	clientReader := bufio.NewReader(conn)
	clientWriter := bufio.NewWriter(conn)

	upstreamReader := bufio.NewReader(upstreamConn)
	upstreamWriter := bufio.NewWriter(upstreamConn)

	return &ProxySession{
		conn:                  conn,
		upstreamConn:          upstreamConn,
		clientRW:              bufio.NewReadWriter(clientReader, clientWriter),
		upstreamRW:            bufio.NewReadWriter(upstreamReader, upstreamWriter),
		onLineReceivedHandler: func(b []byte) []byte { return b }, // Default handler just returns the line unchanged
	}
}

func (p *ProxySession) SetOnLineReceivedHandler(callback func([]byte) []byte) {
	p.onLineReceivedHandler = callback
}

// proxy reads from the 'from' ReaderWriter and writes to 'to' ReaderWriter, and flushes 'to' inside an infinite loop.
// In case of any error, it terminates and returns the error
func (p *ProxySession) proxy(from *bufio.ReadWriter, to *bufio.ReadWriter) error {
	for {
		line, err := from.ReadBytes('\n')
		if err != nil {
			return err
		}
		_, err = to.Write(p.onLineReceivedHandler(line))
		if err != nil {
			return err
		}
		err = to.Flush()
		if err != nil {
			return err
		}
	}
}

// Start starts the proxy, and proxies data between the server and client. This method handles cleanup of both
// connection and the upstream connection. It blocks until either client-proxy or proxy-upstream connection is closed
func (p *ProxySession) Start() {
	slog.Info("start proxy", "client", p.conn.RemoteAddr().String(), "upstream", p.upstreamConn.RemoteAddr())
	defer func() {
		slog.Info("stop proxy", "client", p.conn.RemoteAddr().String(), "upstream", p.upstreamConn.RemoteAddr())
		p.conn.Close()
		p.upstreamConn.Close()
	}()

	var wg sync.WaitGroup

	// Handle client->upstream communication
	wg.Go(func() {
		p.proxy(p.clientRW, p.upstreamRW)
		p.conn.Close()
		p.upstreamConn.Close()
	})

	// Handle upstream->client communication
	wg.Go(func() {
		p.proxy(p.upstreamRW, p.clientRW)
		p.conn.Close()
		p.upstreamConn.Close()
	})

	wg.Wait()
}
