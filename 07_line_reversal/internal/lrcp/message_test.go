package lrcp

import (
	"reflect"
	"testing"
)

func TestParseMessage(t *testing.T) {
	tests := []struct {
		in   string
		want message
		err  bool
	}{
		// Invalid messages
		{"/", message{}, true},
		{"/ack", message{}, true},
		{"connect/", message{}, true},
		{"", message{}, true},
		{"connect", message{}, true},
		{"//connect", message{}, true},
		{"connect//", message{}, true},
		{"///connect///", message{}, true},
		{"/connect/3", message{}, true},
		{"/connect/xyz/", message{}, true},
		{"/connect/-13/", message{}, true},
		{"/connect/33.151/", message{}, true},
		{"/connect//", message{}, true},
		{"/connect/2147483648/", message{}, true},
		{"/connect/2147483647//", message{}, true},
		{"connect/2147483647/", message{}, true},
		{"/close//", message{}, true},
		{"/close/2147483648/", message{}, true},
		{"/close/2147483647//", message{}, true},
		{"close/2147483647/", message{}, true},
		{"close", message{}, true},
		{"ack", message{}, true},
		{"//ack", message{}, true},
		{"ack//", message{}, true},
		{"///ack///", message{}, true},
		{"/ack/3", message{}, true},
		{"/ack/xyz/", message{}, true},
		{"/ack/-13/", message{}, true},
		{"/ack/33.151/", message{}, true},
		{"/ack//", message{}, true},
		{"/ack/2147483648/", message{}, true},
		{"/ack/2147483647/", message{}, true},
		{"/ack/0/", message{}, true},
		{"/ack/123456/", message{}, true},
		{"/ack/0/abc", message{}, true},
		{"/ack/123456/5.13", message{}, true},
		{"/ack/123456/5.13//", message{}, true},
		{"/ack/123456/181", message{}, true},
		{"/ack/123456/-181", message{}, true},
		{"/ack/123456//181/", message{}, true},
		{"/ack/2147483648/2147483648/", message{}, true},
		{"/data/123456/810/hello this is a message", message{}, true},

		{"/connect/2147483647/", message{kind: Connect, sessionId: 2147483647}, false},
		{"/close/2147483647/", message{kind: Close, sessionId: 2147483647}, false},
		{"/connect/0/", message{kind: Connect, sessionId: 0}, false},
		{"/close/0/", message{kind: Close, sessionId: 0}, false},
		{"/connect/123456/", message{kind: Connect, sessionId: 123456}, false},
		{"/close/123456/", message{kind: Close, sessionId: 123456}, false},
		{"/ack/123456/789/", message{kind: Ack, sessionId: 123456, length: 789}, false},
		{"/ack/0/0/", message{kind: Ack, sessionId: 0, length: 0}, false},
		{"/ack/2147483647/2147483647/", message{kind: Ack, sessionId: 2147483647, length: 2147483647}, false},
		{"/data/123456/810/hello this/", message{
			kind:      Data,
			sessionId: 123456,
			pos:       810,
			data:      []byte("hello this"),
		}, false},
		{
			"/data/123456/810//", message{
				kind:      Data,
				sessionId: 123456,
				pos:       810,
				data:      []byte{},
			},
			false,
		},

		// Test escaping
		{
			"/data/123456/810/h/", message{
				kind:      Data,
				sessionId: 123456,
				pos:       810,
				data:      []byte("h"),
			},
			false,
		},
		{
			"/data/123456/810/h\n/", message{
				kind:      Data,
				sessionId: 123456,
				pos:       810,
				data:      []byte("h\n"),
			},
			false,
		},
		{
			`/data/123456/810/foo\/bar\\baz/`, message{
				kind:      Data,
				sessionId: 123456,
				pos:       810,
				data:      []byte(`foo/bar\baz`),
			},
			false,
		},
		// Test more escape sequences
		{
			`/data/0/0/\\\/\\\/\\\//`, message{
				kind:      Data,
				sessionId: 0,
				pos:       0,
				data:      []byte(`\/\/\/`),
			},
			false,
		},
		{
			`/data/100/50/test\\message\\/`, message{
				kind:      Data,
				sessionId: 100,
				pos:       50,
				data:      []byte(`test\message\`),
			},
			false,
		},
		// Test invalid escape sequences
		{`/data/123/456/invalid\x/`, message{}, true},
		{`/data/123/456/invalid\n/`, message{}, true},
		{`/data/1413578440/520/illegal data/has too many/parts/`, message{}, true},
	}

	for _, test := range tests {
		got, err := ParseMessage([]byte(test.in))
		if err == nil && test.err {
			t.Errorf("expected error for ParseMessage(%q)", test.in)
			continue
		}
		if err != nil && !test.err {
			t.Errorf("unexpected error for ParseMessage(%q): %v", test.in, err)
			continue
		}
		if !reflect.DeepEqual(got, test.want) {
			t.Errorf("got %+v, want %+v", got, test.want)
		}
	}
}

func TestMessageSerializer(t *testing.T) {
	tests := []struct {
		msg  message
		want string
	}{
		{message{kind: Connect, sessionId: 0}, "/connect/0/"},
		{message{kind: Connect, sessionId: 123456}, "/connect/123456/"},
		{message{kind: Connect, sessionId: 2147483647}, "/connect/2147483647/"},
		{message{kind: Close, sessionId: 0}, "/close/0/"},
		{message{kind: Close, sessionId: 123456}, "/close/123456/"},
		{message{kind: Close, sessionId: 2147483647}, "/close/2147483647/"},
		{message{kind: Ack, sessionId: 0, length: 0}, "/ack/0/0/"},
		{message{kind: Ack, sessionId: 123456, length: 789}, "/ack/123456/789/"},
		{message{kind: Ack, sessionId: 2147483647, length: 2147483647}, "/ack/2147483647/2147483647/"},
		{message{kind: Data, sessionId: 123456, pos: 810, data: []byte("hello this")}, "/data/123456/810/hello this/"},
		{message{kind: Data, sessionId: 123456, pos: 810, data: []byte{}}, "/data/123456/810//"},
		{message{kind: Data, sessionId: 123456, pos: 810, data: []byte("h")}, "/data/123456/810/h/"},
		{message{kind: Data, sessionId: 123456, pos: 810, data: []byte("h\n")}, "/data/123456/810/h\n/"},
		{message{kind: Data, sessionId: 123456, pos: 810, data: []byte("foo/bar\\baz")}, "/data/123456/810/foo\\/bar\\\\baz/"},
		{message{kind: Data, sessionId: 0, pos: 0, data: []byte("\\/\\/\\/")}, "/data/0/0/\\\\\\/\\\\\\/\\\\\\//"},
		{message{kind: Data, sessionId: 100, pos: 50, data: []byte("test\\message\\")}, "/data/100/50/test\\\\message\\\\/"},
	}

	for _, test := range tests {
		got := string(SerializeMessage(test.msg))
		if got != test.want {
			t.Errorf("MessageSerializer(%+v) = %q, want %q", test.msg, got, test.want)
		}
	}
}
