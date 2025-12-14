package lrcp

import (
	"log/slog"
	"net"
	"sync"
	"time"
)

const maxDataSize = 900 // Max size of data in a segment (in bytes)

// LRCPConn satisfies net.Conn interface for connection oriented protocol
type LRCPConn struct {
	remoteAddr net.Addr
	localAddr  net.Addr
	sessionId  int64

	// Total bytes received is the number of bytes received in this
	// connection.
	totalBytesReceived int64

	mu          sync.Mutex
	recvBuffer  []byte
	waitingRead chan struct{}
	closed      bool

	sendMu          sync.Mutex
	sendBuffer      []byte
	nextPos         int64
	lastSendAckTime time.Time

	unacked []segment

	outbound chan outboundMessage // The same channel the listener uses
}

type segment struct {
	pos     int64
	data    []byte
	payload []byte
}

func (lConn *LRCPConn) Read(b []byte) (int, error) {
	for {
		lConn.mu.Lock()
		if lConn.closed {
			lConn.mu.Unlock()
			return 0, net.ErrClosed
		}

		// We have received bytes in the read buffer
		if len(lConn.recvBuffer) > 0 {
			n := copy(b, lConn.recvBuffer)
			lConn.recvBuffer = lConn.recvBuffer[n:]
			lConn.mu.Unlock()
			return n, nil
		}
		ch := lConn.waitingRead
		lConn.mu.Unlock()
		<-ch
	}
}

func (lConn *LRCPConn) Write(b []byte) (n int, err error) {
	return lConn.handleWrite(b)
}

func (lConn *LRCPConn) Close() error {
	lConn.mu.Lock()
	defer lConn.mu.Unlock()
	lConn.closed = true
	// Wake up any Read() that is blocked
	select {
	case lConn.waitingRead <- struct{}{}:
	default:
	}
	return nil
}

func (lConn *LRCPConn) LocalAddr() net.Addr {
	return lConn.localAddr
}
func (lConn *LRCPConn) RemoteAddr() net.Addr {
	return lConn.remoteAddr
}
func (lConn *LRCPConn) SetDeadline(t time.Time) error {
	return nil
}
func (lConn *LRCPConn) SetReadDeadline(t time.Time) error {
	return nil
}
func (lConn *LRCPConn) SetWriteDeadline(t time.Time) error {
	return nil
}

func (lConn *LRCPConn) appendBytes(b []byte) {
	lConn.mu.Lock()
	defer lConn.mu.Unlock()
	lConn.recvBuffer = append(lConn.recvBuffer, b...)
	// Non blocking, wake up any Read() that is blocked
	select {
	case lConn.waitingRead <- struct{}{}:
	default:
	}
}

func (lConn *LRCPConn) handleConnect() {
	slog.Info("got connect request", "from", lConn.remoteAddr)
	lConn.outbound <- outboundMessage{
		addr: lConn.remoteAddr,
		buffer: SerializeMessage(message{
			kind:      Ack,
			sessionId: lConn.sessionId,
			length:    0,
		}),
	}

	// Start a goroutine that handles unacked segments
	go func() {
		slog.Info("starting unack goroutine")
		ticker := time.NewTicker(retransmissionTimeout)
		defer ticker.Stop()

		for range ticker.C {
			lConn.sendMu.Lock()

			// Check if connection should be closed due to timeout
			// Also check if unacked has some elements
			if len(lConn.unacked) > 0 && time.Now().After(lConn.lastSendAckTime.Add(sessionExpiryTimeout)) {
				slog.Info("client timeout", "addr", lConn.remoteAddr)
				lConn.sendMu.Unlock()
				lConn.Close()
				return
			}

			for _, seg := range lConn.unacked {
				slog.Info("retransmit data", "addr", lConn.remoteAddr, "pos", seg.pos, "buffer", seg.payload)
				lConn.outbound <- outboundMessage{
					addr:   lConn.remoteAddr,
					buffer: seg.payload,
				}
			}
			lConn.sendMu.Unlock()
		}
	}()
}

func (lConn *LRCPConn) handleData(pos int64, data []byte) {
	// 1) We have received all the data upto pos, the current packet can
	// contain data that may exceed totalBytesReceived
	if pos <= lConn.totalBytesReceived {
		// The packet does not contain any new data
		if (pos + int64(len(data))) <= lConn.totalBytesReceived {
			lConn.outbound <- outboundMessage{
				addr: lConn.remoteAddr,
				buffer: SerializeMessage(message{
					kind:      Ack,
					sessionId: lConn.sessionId,
					length:    lConn.totalBytesReceived,
				}),
			}
			return
		}

		// The packet contains new data
		newEnd := pos + int64(len(data))
		startIdx := lConn.totalBytesReceived - pos
		newBytes := data[startIdx:]

		// Send the new bytes received to Read()
		lConn.appendBytes(newBytes)
		lConn.totalBytesReceived = newEnd
		lConn.outbound <- outboundMessage{
			addr: lConn.remoteAddr,
			buffer: SerializeMessage(message{
				kind:      Ack,
				sessionId: lConn.sessionId,
				length:    newEnd,
			}),
		}
		return
	}
	// 2) We have not yet received all data after totalBytesReceived
	lConn.outbound <- outboundMessage{
		addr: lConn.remoteAddr,
		buffer: SerializeMessage(message{
			kind:      Ack,
			sessionId: lConn.sessionId,
			length:    lConn.totalBytesReceived,
		}),
	}
}
func safeSlice(buf []byte, n int) []byte {
	if n > len(buf) {
		n = len(buf)
	}
	return buf[:n]
}

func (lConn *LRCPConn) handleWrite(b []byte) (n int, err error) {
	lConn.sendMu.Lock()
	defer lConn.sendMu.Unlock()
	lConn.sendBuffer = append(lConn.sendBuffer, b...)
	written := len(b)

	// Create segments
	for len(lConn.sendBuffer) > 0 {
		chunk := safeSlice(lConn.sendBuffer, maxDataSize)
		slog.Info("create segment", "chunk", chunk, "pos", lConn.nextPos)
		pos := lConn.nextPos
		lConn.nextPos += int64(len(chunk))
		chunkCpy := make([]byte, len(chunk))
		copy(chunkCpy, chunk)
		seg := segment{
			pos: pos,
			payload: SerializeMessage(message{
				kind:      Data,
				sessionId: lConn.sessionId,
				pos:       pos,
				data:      chunkCpy,
			}),
			data: chunkCpy,
		}

		// Queue for checking retransmission
		lConn.unacked = append(lConn.unacked, seg)

		lConn.outbound <- outboundMessage{
			addr:   lConn.remoteAddr,
			buffer: seg.payload,
		}

		// Drop this chunk from send buffer
		lConn.sendBuffer = lConn.sendBuffer[len(chunk):]
	}
	return written, nil
}

func (lConn *LRCPConn) handleAck(length int64) {
	lConn.sendMu.Lock()
	defer lConn.sendMu.Unlock()
	slog.Info("received ack", "length", length)
	lConn.lastSendAckTime = time.Now()

	if length > lConn.nextPos {
		// Misbehaving client
		lConn.outbound <- outboundMessage{
			addr: lConn.remoteAddr,
			buffer: SerializeMessage(message{
				kind:      Close,
				sessionId: lConn.sessionId,
			}),
		}
		lConn.Close()
	}

	// Drop acked segments
	for len(lConn.unacked) > 0 {
		seg := lConn.unacked[0]
		end := seg.pos + int64(len(seg.data))
		if end <= length {
			// This segment has been acknowledged, drop it
			lConn.unacked = lConn.unacked[1:]
			slog.Info("dropping segment", "end", end, "length", length)
		} else {
			break
		}
	}
}

func (lConn *LRCPConn) sendAck() {
	lConn.sendMu.Lock()
	defer lConn.sendMu.Unlock()
	lConn.outbound <- outboundMessage{
		addr: lConn.remoteAddr,
		buffer: SerializeMessage(message{
			kind:      Ack,
			pos:       lConn.totalBytesReceived,
			sessionId: lConn.sessionId,
		}),
	}
}
