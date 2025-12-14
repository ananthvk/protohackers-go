package lrcp

type MessageType int

const (
	Connect = iota
	Data
	Ack
	Close
)

func (m MessageType) String() string {
	switch m {
	case Connect:
		return "connect"
	case Data:
		return "data"
	case Ack:
		return "ack"
	case Close:
		return "close"
	}
	return "<invalid-type>"
}

type message struct {
	kind      MessageType // Message type
	sessionId int64       // Set when kind=Connect, Data, Ack, Close
	pos       int64       // Set when kind=Data
	data      []byte      // Set when kind=Data
	length    int64       // Set when kind=Ack
}

var (
	connectBytes = []byte("connect")
	dataBytes    = []byte("data")
	ackBytes     = []byte("ack")
	closeBytes   = []byte("close")
)
