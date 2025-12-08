package internal

import "testing"

func TestReplaceBoguscoin(t *testing.T) {
	const targetBogusCoin = "[BOGUS]"
	tests := []struct {
		in   string
		want string
	}{
		// Test replacement, address is the message
		{"7F1u3wSD5RbOHQmupo9nx4TnhQ", "[BOGUS]"},

		// Test replacement with trailing spaces
		{" 7F1u3wSD5RbOHQmupo9nx4TnhQ", " [BOGUS]"},
		{"7F1u3wSD5RbOHQmupo9nx4TnhQ ", "[BOGUS] "},
		{"   7F1u3wSD5RbOHQmupo9nx4TnhQ    ", "   [BOGUS]    "},

		// Test replacement with newline characters
		{"\n7F1u3wSD5RbOHQmupo9nx4TnhQ\n", "\n[BOGUS]\n"},
		{"7F1u3wSD5RbOHQmupo9nx4TnhQ \n  ", "[BOGUS] \n  "},
		{"\n7F1u3wSD5RbOHQmupo9nx4TnhQ\n\n\n", "\n[BOGUS]\n\n\n"},

		{"Hi alice, please send payment to 7YWHMfk9JZe0LM0g1ZauHuiSxhI", "Hi alice, please send payment to [BOGUS]"},

		// Valid boguscoin addresses
		{"7F1u3wSD5RbOHQmupo9nx4TnhQ", "[BOGUS]"},
		{"7iKDZEwPZSqIvDnHvVN2r0hUWXD5rHX", "[BOGUS]"},
		{"7LOrwbDlS8NujgjddyogWgIM93MV5N2VR", "[BOGUS]"},
		{"7adNeSwJkMakpEcln9HEtthSRtxdmEHOT8T", "[BOGUS]"},

		// Invalid boguscoin addresses (does not start with 7)
		{"6F1u3wSD5RbOHQmupo9nx4TnhQ", "6F1u3wSD5RbOHQmupo9nx4TnhQ"},
		{"This is a AiKDZEwPZSqIvDnHvVN2r0hUWXD5rHX\n", "This is a AiKDZEwPZSqIvDnHvVN2r0hUWXD5rHX\n"},
		{"\n\tkLOrwbDlS8NujgjddyogWgIM93MV5N2VR", "\n\tkLOrwbDlS8NujgjddyogWgIM93MV5N2VR"},
		{"\t8adNeSwJkMakpEcln9HEtthSRtxdmEHOT8T\n\n", "\t8adNeSwJkMakpEcln9HEtthSRtxdmEHOT8T\n\n"},

		// Invalid boguscoin addresses (does not contain only alphanumeric)
		{"6F1u3wSD_RbOHQmupo9nx4TnhQ", "6F1u3wSD_RbOHQmupo9nx4TnhQ"},
		{"AiKDZEwP-SqIvDnHvVN2r0h UWXD5rHX", "AiKDZEwP-SqIvDnHvVN2r0h UWXD5rHX"},
		{"kLOrwbDlS8NujgjddyoggI \n M93MV5N2VR", "kLOrwbDlS8NujgjddyoggI \n M93MV5N2VR"},
		{"$adNeSwJkMakpEcln9HEtthSRtxdmEHOT8T", "$adNeSwJkMakpEcln9HEtthSRtxdmEHOT8T"},

		// Invalid boguscoin addresses (invalid length)
		{"", ""},
		{"     ", "     "},
		{"7XXXXXXXXXXXXXXXXXXXXXXXX", "7XXXXXXXXXXXXXXXXXXXXXXXX"},                       // 25 chars
		{"7XXXXXXXXXXXXXXXXXXXXXXXXXXXXXXXXXXX", "7XXXXXXXXXXXXXXXXXXXXXXXXXXXXXXXXXXX"}, // 36 chars

		// Valid boguscoin (boundary 26 & 35 chars)
		{"7XXXXXXXXXXXXXXXXXXXXXXXXX", "[BOGUS]"},          // 26 chars
		{"7XXXXXXXXXXXXXXXXXXXXXXXXXXXXXXXXXX", "[BOGUS]"}, // 35 chars

		{"[PoorCoder99] 7xQMlMvFkfzMx4OFlbLEvfwyDrOKmSpReLO\n", "[PoorCoder99] [BOGUS]\n"},
		{"[PoorCoder99] 7xQMlMvFkfzMx4OFlbLEvfwyDrOKmSpReLO", "[PoorCoder99] [BOGUS]"},
		{"[PoorCoder99] 7xQMlMvFkfzMx4OFlbLEvfwyDrOKmSpR-xa", "[PoorCoder99] 7xQMlMvFkfzMx4OFlbLEvfwyDrOKmSpR-xa"},
	}
	for _, test := range tests {
		got := ReplaceBoguscoin(test.in, targetBogusCoin)
		if got != test.want {
			t.Errorf("got %q, want %q for input %q", got, test.want, test.in)
		}
	}
}
