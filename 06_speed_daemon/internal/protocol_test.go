package internal

import (
	"bytes"
	"reflect"
	"testing"
)

func TestReadPlate(t *testing.T) {
	t.Run("testplate1", func(t *testing.T) {
		b := []byte{0x04, 0x55, 0x4e, 0x31, 0x58, 0x00, 0x00, 0x03, 0xe8}
		r := bytes.NewReader(b)
		plate, err := readPlate(r)
		if err != nil {
			t.Errorf("unexpected error: %v", err)
			return
		}
		want := PlateMessage{plate: "UN1X", timestamp: 1000}
		if plate != want {
			t.Errorf("want %+v, got %+v", want, plate)
		}
	})
	t.Run("testplate-2", func(t *testing.T) {
		b := []byte{0x07, 0x52, 0x45, 0x30, 0x35, 0x42, 0x4b, 0x47, 0x00, 0x01, 0xe2, 0x40}
		r := bytes.NewReader(b)
		plate, err := readPlate(r)
		if err != nil {
			t.Errorf("unexpected error: %v", err)
			return
		}
		want := PlateMessage{plate: "RE05BKG", timestamp: 123456}
		if plate != want {
			t.Errorf("want %+v, got %+v", want, plate)
		}
	})
}

func TestReadWantHeartBeat(t *testing.T) {
	t.Run("test-1", func(t *testing.T) {
		b := []byte{0x00, 0x00, 0x00, 0x0a}
		r := bytes.NewReader(b)
		heartbeat, err := readWantHeartbeat(r)
		if err != nil {
			t.Errorf("unexpected error: %v", err)
			return
		}
		want := WantHeartbeatMessage{interval: 10}
		if heartbeat != want {
			t.Errorf("want %+v, got %+v", want, heartbeat)
		}
	})
	t.Run("test-2", func(t *testing.T) {
		b := []byte{0x00, 0x00, 0x04, 0xdb}
		r := bytes.NewReader(b)
		heartbeat, err := readWantHeartbeat(r)
		if err != nil {
			t.Errorf("unexpected error: %v", err)
			return
		}
		want := WantHeartbeatMessage{interval: 1243}
		if heartbeat != want {
			t.Errorf("want %+v, got %+v", want, heartbeat)
		}
	})
}

func TestIAmCamera(t *testing.T) {
	t.Run("test-1", func(t *testing.T) {
		b := []byte{0x00, 0x42, 0x00, 0x64, 0x00, 0x3c}
		r := bytes.NewReader(b)
		message, err := readIAmCamera(r)
		if err != nil {
			t.Errorf("unexpected error: %v", err)
			return
		}
		want := IAmCameraMessage{road: 66, mile: 100, limit: 60}
		if message != want {
			t.Errorf("want %+v, got %+v", want, message)
		}
	})
	t.Run("test-2", func(t *testing.T) {
		b := []byte{0x01, 0x70, 0x04, 0xd2, 0x00, 0x28}
		r := bytes.NewReader(b)
		message, err := readIAmCamera(r)
		if err != nil {
			t.Errorf("unexpected error: %v", err)
			return
		}
		want := IAmCameraMessage{road: 368, mile: 1234, limit: 40}
		if message != want {
			t.Errorf("want %+v, got %+v", want, message)
		}
	})
}

func TestIAmDispatcher(t *testing.T) {
	t.Run("test-1", func(t *testing.T) {
		b := []byte{0x01, 0x00, 0x42}
		r := bytes.NewReader(b)
		message, err := readIAmDispatcher(r)
		if err != nil {
			t.Errorf("unexpected error: %v", err)
			return
		}
		want := IAmDispatcherMessage{roads: []uint16{66}}
		if !reflect.DeepEqual(message, want) {
			t.Errorf("want %+v, got %+v", want, message)
		}
	})
	t.Run("test-2", func(t *testing.T) {
		b := []byte{0x03, 0x00, 0x42, 0x01, 0x70, 0x13, 0x88}
		r := bytes.NewReader(b)
		message, err := readIAmDispatcher(r)
		if err != nil {
			t.Errorf("unexpected error: %v", err)
			return
		}
		want := IAmDispatcherMessage{roads: []uint16{66, 368, 5000}}
		if !reflect.DeepEqual(message, want) {
			t.Errorf("want %+v, got %+v", want, message)
		}
	})
}

func TestReadMessage(t *testing.T) {
	t.Run("plate message", func(t *testing.T) {
		b := []byte{0x20, 0x04, 0x55, 0x4e, 0x31, 0x58, 0x00, 0x00, 0x03, 0xe8}
		r := bytes.NewReader(b)
		message, err := ReadMessage(r)
		if err != nil {
			t.Errorf("unexpected error: %v", err)
			return
		}
		want := PlateMessage{plate: "UN1X", timestamp: 1000}
		if message != want {
			t.Errorf("want %+v, got %+v", want, message)
		}
	})

	t.Run("want heartbeat message", func(t *testing.T) {
		b := []byte{0x40, 0x00, 0x00, 0x00, 0x0a}
		r := bytes.NewReader(b)
		message, err := ReadMessage(r)
		if err != nil {
			t.Errorf("unexpected error: %v", err)
			return
		}
		want := WantHeartbeatMessage{interval: 10}
		if message != want {
			t.Errorf("want %+v, got %+v", want, message)
		}
	})

	t.Run("i am camera message", func(t *testing.T) {
		b := []byte{0x80, 0x00, 0x42, 0x00, 0x64, 0x00, 0x3c}
		r := bytes.NewReader(b)
		message, err := ReadMessage(r)
		if err != nil {
			t.Errorf("unexpected error: %v", err)
			return
		}
		want := IAmCameraMessage{road: 66, mile: 100, limit: 60}
		if message != want {
			t.Errorf("want %+v, got %+v", want, message)
		}
	})

	t.Run("i am dispatcher message", func(t *testing.T) {
		b := []byte{0x81, 0x01, 0x00, 0x42}
		r := bytes.NewReader(b)
		message, err := ReadMessage(r)
		if err != nil {
			t.Errorf("unexpected error: %v", err)
			return
		}
		want := IAmDispatcherMessage{roads: []uint16{66}}
		if !reflect.DeepEqual(message, want) {
			t.Errorf("want %+v, got %+v", want, message)
		}
	})

	t.Run("unknown message type", func(t *testing.T) {
		b := []byte{0xFF, 0x00, 0x00}
		r := bytes.NewReader(b)
		_, err := ReadMessage(r)
		if err == nil {
			t.Error("expected error for unknown message type, got nil")
		}
	})
}

func TestWriteError(t *testing.T) {
	t.Run("test1", func(t *testing.T) {
		var buf bytes.Buffer
		err := WriteError(&buf, "bad")
		if err != nil {
			t.Errorf("unexpected error: %v", err)
			return
		}
		want := []byte{0x10, 0x03, 0x62, 0x61, 0x64}
		got := buf.Bytes()
		if !bytes.Equal(got, want) {
			t.Errorf("want %x, got %x", want, got)
		}
	})

	t.Run("test2", func(t *testing.T) {
		var buf bytes.Buffer
		err := WriteError(&buf, "illegal msg")
		if err != nil {
			t.Errorf("unexpected error: %v", err)
			return
		}
		want := []byte{0x10, 0x0b, 0x69, 0x6c, 0x6c, 0x65, 0x67, 0x61, 0x6c, 0x20, 0x6d, 0x73, 0x67}
		got := buf.Bytes()
		if !bytes.Equal(got, want) {
			t.Errorf("want %x, got %x", want, got)
		}
	})
}

func TestWriteTicket(t *testing.T) {
	t.Run("test1", func(t *testing.T) {
		var buf bytes.Buffer
		ticket := TicketMessage{
			plate:      "UN1X",
			road:       66,
			mile1:      100,
			timestamp1: 123456,
			mile2:      110,
			timestamp2: 123816,
			speed:      10000,
		}
		err := WriteTicket(&buf, ticket)
		if err != nil {
			t.Errorf("unexpected error: %v", err)
			return
		}
		want := []byte{0x21, 0x04, 0x55, 0x4e, 0x31, 0x58, 0x00, 0x42, 0x00, 0x64, 0x00, 0x01, 0xe2, 0x40, 0x00, 0x6e, 0x00, 0x01, 0xe3, 0xa8, 0x27, 0x10}
		got := buf.Bytes()
		if !bytes.Equal(got, want) {
			t.Errorf("want %x, got %x", want, got)
		}
	})

	t.Run("test2", func(t *testing.T) {
		var buf bytes.Buffer
		ticket := TicketMessage{
			plate:      "RE05BKG",
			road:       368,
			mile1:      1234,
			timestamp1: 1000000,
			mile2:      1235,
			timestamp2: 1000060,
			speed:      6000,
		}
		err := WriteTicket(&buf, ticket)
		if err != nil {
			t.Errorf("unexpected error: %v", err)
			return
		}
		want := []byte{0x21, 0x07, 0x52, 0x45, 0x30, 0x35, 0x42, 0x4b, 0x47, 0x01, 0x70, 0x04, 0xd2, 0x00, 0x0f, 0x42, 0x40, 0x04, 0xd3, 0x00, 0x0f, 0x42, 0x7c, 0x17, 0x70}
		got := buf.Bytes()
		if !bytes.Equal(got, want) {
			t.Errorf("want %x, got %x", want, got)
		}
	})
}

func TestWriteHeartbeat(t *testing.T) {
	var buf bytes.Buffer
	err := WriteHeartbeat(&buf)
	if err != nil {
		t.Errorf("unexpected error: %v", err)
		return
	}
	want := []byte{0x41}
	got := buf.Bytes()
	if !bytes.Equal(got, want) {
		t.Errorf("want %x, got %x", want, got)
	}
}
