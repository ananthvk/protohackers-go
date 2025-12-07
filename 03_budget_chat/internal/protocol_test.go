package internal

import "testing"

func TestValidateName(t *testing.T) {
	tests := []struct {
		in   string
		want bool
	}{
		{" ", false},
		{"xyz123", true},
		{"12345", false},
		{"A123", true},
		{"", false},
		{"X", true},
		{"x", true},
		{"1", false},
		{"1x", true},
		{"abCd12", true},
		{"ab##", false},
		{"abcef", true},
		{"AB$$12", false},
	}
	for _, test := range tests {
		got := validateName(test.in)
		if got != test.want {
			t.Errorf("got %v, want %v for input %q", got, test.want, test.in)
		}
	}
}

func TestFormatBroadcast(t *testing.T) {
	tests := []struct {
		name string
		msg  string
		want string
	}{
		{"alice", "hello world", "[alice] hello world"},
		{"bob", "how are you?", "[bob] how are you?"},
		{"user123", "", "[user123] "},
		{"john", "multiple words here", "[john] multiple words here"},
		{"A", "x", "[A] x"},
	}
	for _, test := range tests {
		got := formatBroadcast(test.name, test.msg)
		if got != test.want {
			t.Errorf("formatBroadcast(%q, %q) = %q, want %q", test.name, test.msg, got, test.want)
		}
	}
}

func TestFormatUserList(t *testing.T) {
	tests := []struct {
		users []string
		want  string
	}{
		{[]string{"alice", "bob", "charlie"}, "* The room contains: alice, bob, charlie"},
		{[]string{"alice"}, "* The room contains: alice"},
		{[]string{}, "* The room contains: "},
		{[]string{"user1", "user2"}, "* The room contains: user1, user2"},
		{[]string{"A", "B", "C", "D"}, "* The room contains: A, B, C, D"},
	}
	for _, test := range tests {
		got := formatUserList(test.users)
		if got != test.want {
			t.Errorf("formatUserList(%v) = %q, want %q", test.users, got, test.want)
		}
	}
}

func TestFormatNotification(t *testing.T) {
	tests := []struct {
		username string
		message  string
		want     string
	}{
		{"alice", "has joined", "* alice: has joined"},
		{"bob", "has left", "* bob: has left"},
		{"user123", "", "* user123: "},
		{"john", "is typing", "* john: is typing"},
		{"A", "x", "* A: x"},
	}
	for _, test := range tests {
		got := formatNotification(test.username, test.message)
		if got != test.want {
			t.Errorf("formatNotification(%q, %q) = %q, want %q", test.username, test.message, got, test.want)
		}
	}
}
