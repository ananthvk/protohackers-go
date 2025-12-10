package internal

import (
	"log/slog"
	"time"
)

func StartWriteLoop(connState *ConnState) {
	slog.Info("started writer loop", "client", connState.conn.RemoteAddr().String())
	client := connState.conn.RemoteAddr().String()
	var heartbeatTicker *time.Ticker
	if connState.heartbeat != 0 {
		// Set the ticker
		slog.Info("setting initial heartbeat interval", "interval", connState.heartbeat, "client", client)
		heartbeatTicker = time.NewTicker(connState.heartbeat)
	}
	for {
		var hbC <-chan time.Time
		if heartbeatTicker != nil {
			hbC = heartbeatTicker.C
		}
		select {
		case hbDuration := <-connState.heartbeatControl:
			slog.Info("received heartbeat control message", "new_interval", hbDuration, "client", client)
			if heartbeatTicker != nil {
				heartbeatTicker.Stop()
				heartbeatTicker = nil
			}
			if hbDuration != 0 {
				heartbeatTicker = time.NewTicker(hbDuration)
			}
		case ticket := <-connState.outgoing:
			slog.Info("dispatching ticket", "ticket", ticket, "client", client)
			err := WriteTicket(connState.conn, TicketMessage{
				plate:      string(ticket.plate),
				road:       uint16(ticket.road),
				mile1:      ticket.mile1,
				mile2:      ticket.mile2,
				timestamp1: ticket.timestamp1,
				timestamp2: ticket.timestamp2,
				speed:      ticket.speed,
			})
			if err != nil {
				slog.Error("dispatch ticket failed", "error", err, "client", client)
				return
			}
		case <-connState.kill:
			slog.Info("killed writer loop", "client", client)
			return
		case <-hbC:
			err := WriteHeartbeat(connState.conn)
			if err != nil {
				slog.Error("heartbeat send failed", "error", err, "client", client)
				return
			}
		}
	}
}
