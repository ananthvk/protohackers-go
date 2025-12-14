package lrcp

// LRCPAddress satisfies net.Addr interface
type LRCPAddress struct {
	address string
}

func (l *LRCPAddress) Network() string {
	return "lrcp"
}

func (l *LRCPAddress) String() string {
	return l.address
}
