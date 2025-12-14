package lrcp

import (
	"context"
	"errors"
	"log/slog"
	"net"
	"sync"
	"time"
)

const (
	maxPacketSize         = 1000
	retransmissionTimeout = time.Second * 3
	sessionExpiryTimeout  = time.Second * 60
	outboundQueueSize     = 32
)

type outboundMessage struct {
	buffer []byte
	addr   net.Addr
}

type LRCPListener struct {
	addr          *LRCPAddress
	conn          net.PacketConn
	sessions      map[int64]*LRCPConn
	sessionMu     sync.Mutex
	newConnection chan *LRCPConn
	outbound      chan outboundMessage
}

type ListenConfig struct {
}

func (listener *LRCPListener) Accept() (net.Conn, error) {
	conn, ok := <-listener.newConnection
	if !ok {
		return nil, net.ErrClosed
	}
	return conn, nil
}

func (listener *LRCPListener) Close() error {
	err := listener.conn.Close()
	close(listener.newConnection)
	close(listener.outbound)
	return err
}

func (listener *LRCPListener) Addr() net.Addr {
	return listener.addr
}

func (lc *ListenConfig) Listen(ctx context.Context, network, address string) (net.Listener, error) {
	if network != "lrcp" {
		return nil, errors.New("invalid network type, needs to be 'lrcp'")
	}
	listenerConfig := net.ListenConfig{}
	udpConn, err := listenerConfig.ListenPacket(ctx, "udp", address)
	if err != nil {
		return nil, err
	}
	listener := &LRCPListener{
		addr:          &LRCPAddress{address: address},
		conn:          udpConn,
		sessions:      map[int64]*LRCPConn{},
		newConnection: make(chan *LRCPConn),
		outbound:      make(chan outboundMessage, outboundQueueSize),
	}

	// Start the reader loop
	var buffer [maxPacketSize]byte
	go func() {
		for {
			n, fromAddr, err := udpConn.ReadFrom(buffer[:])
			if n > 0 {
				listener.handleMessage(fromAddr, buffer[:n])
			}
			if err != nil {
				slog.Error("read error", "error", err)
				return
			}
		}
	}()
	// Start the writer loop
	go func() {
		for {
			msg, ok := <-listener.outbound
			if !ok {
				return
			}
			udpConn.WriteTo(msg.buffer, msg.addr)
		}
	}()
	return listener, nil
}

func (l *LRCPListener) getSession(sessionId int64) (*LRCPConn, bool) {
	l.sessionMu.Lock()
	defer l.sessionMu.Unlock()
	conn, ok := l.sessions[sessionId]
	return conn, ok
}

func (l *LRCPListener) setSession(sessionId int64, conn *LRCPConn) {
	l.sessionMu.Lock()
	defer l.sessionMu.Unlock()
	l.sessions[sessionId] = conn
}

func (l *LRCPListener) deleteSession(sessionId int64) {
	l.sessionMu.Lock()
	defer l.sessionMu.Unlock()
	delete(l.sessions, sessionId)
}

func (listener *LRCPListener) sendClose(sessionId int64, addr net.Addr) {
	listener.outbound <- outboundMessage{
		addr: addr,
		buffer: SerializeMessage(message{
			kind:      Close,
			sessionId: sessionId,
		}),
	}
}

func (listener *LRCPListener) handleMessage(fromAddr net.Addr, buffer []byte) {
	msg, err := ParseMessage(buffer)
	slog.Info("UDP", "from", fromAddr, "buffer", string(buffer), "err", err)
	if err != nil {
		return
	}

	switch msg.kind {

	case Connect:
		conn, ok := listener.getSession(msg.sessionId)
		if ok {
			conn.sendAck()
			return
		}

		conn = &LRCPConn{
			remoteAddr:      fromAddr,
			localAddr:       listener.conn.LocalAddr(),
			sessionId:       msg.sessionId,
			waitingRead:     make(chan struct{}),
			outbound:        listener.outbound,
			lastSendAckTime: time.Now(),
		}
		listener.setSession(msg.sessionId, conn)

		conn.handleConnect()
		listener.newConnection <- conn
		return

	case Data:
		conn, ok := listener.getSession(msg.sessionId)
		if !ok {
			listener.sendClose(msg.sessionId, fromAddr)
			return
		}
		conn.handleData(msg.pos, msg.data)
		return

	case Ack:
		conn, ok := listener.getSession(msg.sessionId)
		if !ok {
			listener.sendClose(msg.sessionId, fromAddr)
			return
		}
		conn.handleAck(msg.length)
		return

	case Close:
		_, ok := listener.getSession(msg.sessionId)
		if !ok {
			listener.sendClose(msg.sessionId, fromAddr)
			return
		}

		// conn.handleClose()
		listener.sendClose(msg.sessionId, fromAddr)
		listener.deleteSession(msg.sessionId)
		return
	}
}
