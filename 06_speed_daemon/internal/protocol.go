package internal

import (
	"encoding/binary"
	"errors"
	"io"
)

type Message any

// Server messages (These messages are only sent by the server)
// ================

// 0x10
type ErrorMessage struct {
	msg string
}

// 0x21
type TicketMessage struct {
	plate      string
	road       uint16
	mile1      uint16
	timestamp1 uint32
	mile2      uint16
	timestamp2 uint32
	speed      uint16 // x100 miles per hour
}

// 0x41
// Heartbeat

// Client messages (These messages are only sent by the client)
// ================

// 0x20
type PlateMessage struct {
	plate     string
	timestamp uint32
}

// 0x40
type WantHeartbeatMessage struct {
	interval uint32 // in seconds
}

// 0x80
type IAmCameraMessage struct {
	road  uint16
	mile  uint16
	limit uint16
}

// 0x81
type IAmDispatcherMessage struct {
	roads []uint16
}

var (
	ErrInvalidMessageType = errors.New("invalid message type")
)

func ReadMessage(r io.Reader) (Message, error) {
	// Read the type information (first byte)
	var typ [1]byte
	if _, err := io.ReadFull(r, typ[:]); err != nil {
		return nil, err
	}
	switch typ[0] {
	case 0x20: // Plate
		return readPlate(r)
	case 0x40: // WantHeartbeat
		return readWantHeartbeat(r)
	case 0x80: // IAmCamera
		return readIAmCamera(r)
	case 0x81: // IAmDispatcher
		return readIAmDispatcher(r)
	}
	return nil, ErrInvalidMessageType
}

func readLengthPrefixedString(r io.Reader) (string, error) {
	// Read the length of the string
	var length [1]byte
	if _, err := io.ReadFull(r, length[:]); err != nil {
		return "", err
	}
	// Read the string bytes
	buffer := make([]byte, length[0])
	if _, err := io.ReadFull(r, buffer[:]); err != nil {
		return "", err
	}
	return string(buffer), nil
}

func readPlate(r io.Reader) (PlateMessage, error) {
	plate, err := readLengthPrefixedString(r)
	if err != nil {
		return PlateMessage{}, err
	}
	var timestamp uint32
	if err := binary.Read(r, binary.BigEndian, &timestamp); err != nil {
		return PlateMessage{}, err
	}
	return PlateMessage{plate: plate, timestamp: timestamp}, nil
}

func readWantHeartbeat(r io.Reader) (WantHeartbeatMessage, error) {
	var interval uint32
	if err := binary.Read(r, binary.BigEndian, &interval); err != nil {
		return WantHeartbeatMessage{}, err
	}
	return WantHeartbeatMessage{interval: interval}, nil
}

func readIAmCamera(r io.Reader) (IAmCameraMessage, error) {
	message := IAmCameraMessage{}
	if err := binary.Read(r, binary.BigEndian, &message.road); err != nil {
		return IAmCameraMessage{}, err
	}
	if err := binary.Read(r, binary.BigEndian, &message.mile); err != nil {
		return IAmCameraMessage{}, err
	}
	if err := binary.Read(r, binary.BigEndian, &message.limit); err != nil {
		return IAmCameraMessage{}, err
	}
	return message, nil
}
func readIAmDispatcher(r io.Reader) (IAmDispatcherMessage, error) {
	var numRoads [1]byte
	if _, err := io.ReadFull(r, numRoads[:]); err != nil {
		return IAmDispatcherMessage{}, err
	}
	roads := make([]uint16, numRoads[0])
	for i := range roads {
		if err := binary.Read(r, binary.BigEndian, &roads[i]); err != nil {
			return IAmDispatcherMessage{}, err
		}
	}
	return IAmDispatcherMessage{roads: roads}, nil
}

// WriteError writes the error as a length prefixed string to the stream.
func WriteError(w io.Writer, msg string) error {
	// Write message type (0x10)
	if _, err := w.Write([]byte{0x10}); err != nil {
		return err
	}
	// Write length-prefixed string
	if len(msg) > 255 {
		return errors.New("error message too long")
	}
	if _, err := w.Write([]byte{byte(len(msg))}); err != nil {
		return err
	}
	_, err := w.Write([]byte(msg))
	return err
}

func WriteTicket(w io.Writer, ticket TicketMessage) error {
	// Write message type (0x21)
	if _, err := w.Write([]byte{0x21}); err != nil {
		return err
	}
	// Write length-prefixed plate string
	if len(ticket.plate) > 255 {
		return errors.New("plate string too long")
	}
	if _, err := w.Write([]byte{byte(len(ticket.plate))}); err != nil {
		return err
	}
	if _, err := w.Write([]byte(ticket.plate)); err != nil {
		return err
	}
	// Write the remaining fields in big-endian format
	if err := binary.Write(w, binary.BigEndian, ticket.road); err != nil {
		return err
	}
	if err := binary.Write(w, binary.BigEndian, ticket.mile1); err != nil {
		return err
	}
	if err := binary.Write(w, binary.BigEndian, ticket.timestamp1); err != nil {
		return err
	}
	if err := binary.Write(w, binary.BigEndian, ticket.mile2); err != nil {
		return err
	}
	if err := binary.Write(w, binary.BigEndian, ticket.timestamp2); err != nil {
		return err
	}
	return binary.Write(w, binary.BigEndian, ticket.speed)
}

func WriteHeartbeat(w io.Writer) error {
	var b [1]byte
	b[0] = 0x41
	_, err := w.Write(b[:])
	return err
}
