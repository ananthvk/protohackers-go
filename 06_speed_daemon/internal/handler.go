package internal

import (
	"errors"
	"log/slog"
	"net"
	"time"
)

const queueSize = 100

// Handle handles a single client connection. This should be run in a separate gorutine so that requests can be handled
// concurrently.
func Handle(speedServer *SpeedServer, connection net.Conn) {
	client := connection.RemoteAddr().String()
	slog.Info("client connected", "address", client)
	isCamera := false
	isDispatcher := false
	isHeartbeatInitialized := false
	var cameraDetails IAmCameraMessage
	connState := &ConnState{
		conn:             connection,
		outgoing:         make(chan Ticket, queueSize),
		heartbeat:        0,
		kill:             make(chan struct{}),
		heartbeatControl: make(chan time.Duration),
	}

	defer func() {
		slog.Info("client disconnecting", "address", client)
		connState.kill <- struct{}{}
		close(connState.kill)
		if isDispatcher {
			speedServer.dispatchers.RemoveConn(connState)
		}
		connection.Close()
		slog.Info("client disconnected", "address", client)
	}()

	// Start writer loop
	go StartWriteLoop(connState)

	for {
		message, err := ReadMessage(connection)
		if err != nil {
			if errors.Is(err, ErrInvalidMessageType) {
				WriteError(connection, "Invalid message type")
			}
			return
		}
		switch v := message.(type) {
		case WantHeartbeatMessage:
			slog.Info("client initialized heartbeat", "message", v, "client", client)
			if isHeartbeatInitialized {
				slog.Info("client error", "reason", "client already set heartbeat", "client", client)
				WriteError(connection, "client has already set a heartbeat interval on this connection")
				return
			}
			isHeartbeatInitialized = true
			interval := time.Duration(v.interval) * time.Second / 10.0
			if v.interval != 0 {
				connState.heartbeatControl <- interval
			}
			slog.Info("initialized heartbeat", "interval", interval, "client", client)
		case IAmCameraMessage:
			if isCamera || isDispatcher {
				slog.Info("client error", "reason", "client has already identified as dispatcher/camera", "client", client)
				WriteError(connection, "client has already identified as a dispatcher/camera")
				return
			}
			isCamera = true
			cameraDetails = v
			speedServer.store.SetLimit(Road(cameraDetails.road), cameraDetails.limit)
			slog.Info("client identification", "type", "camera", "client", client)
		case IAmDispatcherMessage:
			if isCamera || isDispatcher {
				slog.Info("client error", "reason", "client has already identified as dispatcher/camera", "client", client)
				WriteError(connection, "client has already identified as a dispatcher/camera")
				return
			}
			speedServer.dispatchers.AddConn(connState, v.roads)
			isDispatcher = true
			slog.Info("client identification", "type", "dispatcher", "client", client)
			// Check if there are any pending tickets that need to be sent
			for _, road := range v.roads {
				pending := speedServer.store.GetPending(road)
				slog.Info("dispatching pending tickets", "road", road, "count", len(pending), "client", client)
				for _, p := range pending {
					connState.outgoing <- p
				}
			}
		case PlateMessage:
			if !isCamera {
				slog.Info("client error", "reason", "only camera can send plate message", "client", client)
				WriteError(connection, "only camera can send plate message")
				return
			}
			ticket := speedServer.store.AddObservation(PlateObservation{
				timestamp: v.timestamp,
				plate:     Plate(v.plate),
				road:      Road(cameraDetails.road),
				mile:      cameraDetails.mile,
			})
			if ticket != nil {
				slog.Info("ticket generated", "ticket", ticket, "client", client)
				dipatcherAddr := speedServer.dispatchers.DispatchTicket(*ticket)
				if dipatcherAddr != "" {
					slog.Info("ticket sent to dispatcher", "ticket", ticket, "client", client, "dispatcher_client", dipatcherAddr)
				} else {
					slog.Info("ticket marked pending", "ticket", ticket, "client", client)
					speedServer.store.AddPending(*ticket)
				}
			}
		default:
			panic("Invalid type returned from ReadMessage")
		}
	}
}
