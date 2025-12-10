package internal

import (
	"net"
	"sync"
	"time"
)

type ConnState struct {
	conn             net.Conn
	outgoing         chan Ticket        // nil until dispatcher identifies itself
	heartbeat        time.Duration      // 0 until WantHeartbeat message
	kill             chan struct{}      // When a message is sent on this channel, the writer loop + connection is closed
	heartbeatControl chan time.Duration // Send a time.Duration on this channel to enable heartbeats
}

type SpeedServer struct {
	store       *Store
	dispatchers *DispatcherHub
}
type DispatcherHub struct {
	mu              sync.Mutex
	roadDispatchers map[uint16]map[*ConnState]struct{}
	dispatcherRoads map[*ConnState]map[uint16]struct{}
}

func (d *DispatcherHub) AddConn(connState *ConnState, roads []uint16) {
	d.mu.Lock()
	defer d.mu.Unlock()
	for _, r := range roads {
		if d.roadDispatchers[r] == nil {
			d.roadDispatchers[r] = make(map[*ConnState]struct{})
		}
		d.roadDispatchers[r][connState] = struct{}{}
		if d.dispatcherRoads[connState] == nil {
			d.dispatcherRoads[connState] = make(map[uint16]struct{})
		}
		d.dispatcherRoads[connState][r] = struct{}{}
	}
}

func (d *DispatcherHub) RemoveConn(connState *ConnState) {
	d.mu.Lock()
	defer d.mu.Unlock()
	roads := d.dispatcherRoads[connState]
	for r := range roads {
		delete(d.roadDispatchers[r], connState)
		if len(d.roadDispatchers[r]) == 0 {
			delete(d.roadDispatchers, r)
		}
	}
	delete(d.dispatcherRoads, connState)
	close(connState.outgoing)
}

func (d *DispatcherHub) DispatchTicket(message Ticket) string {
	d.mu.Lock()
	defer d.mu.Unlock()
	dispatchers := d.roadDispatchers[uint16(message.road)]
	// Only dispatch to one dispatcher
	for conn := range dispatchers {
		sent := false
		if conn.outgoing == nil {
			continue
		}
		select {
		case conn.outgoing <- message:
			sent = true
		default:
		}
		if sent {
			return conn.conn.RemoteAddr().String()
		}
	}
	return ""
}

func NewSpeedServer() *SpeedServer {
	return &SpeedServer{
		store: NewStore(),
		dispatchers: &DispatcherHub{
			roadDispatchers: map[uint16]map[*ConnState]struct{}{},
			dispatcherRoads: map[*ConnState]map[uint16]struct{}{},
		},
	}
}
