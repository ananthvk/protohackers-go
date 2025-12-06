package internal

import "testing"

func TestParseMessage(t *testing.T) {
	tests := []struct {
		name string
		in   []byte
		want Message
		err  bool
	}{
		{
			name: "valid insert",
			in:   []byte{'I', 0, 0, 0, 10, 0, 0, 0, 50}, // I, ts=10, price=50
			want: Message{messageType: 'I', field1: 10, field2: 50},
		},
		{
			name: "valid query",
			in:   []byte{'Q', 0, 0, 0, 10, 0, 0, 0, 20}, // Q, min=10, max=20
			want: Message{messageType: 'Q', field1: 10, field2: 20},
		},
		{
			name: "invalid type",
			in:   []byte{'X', 0, 0, 0, 1, 0, 0, 0, 2},
			err:  true,
		},
		{
			name: "too short",
			in:   []byte{1, 2, 3},
			err:  true,
		},
		{
			name: "too long",
			in:   append([]byte{'I'}, make([]byte, 20)...),
			err:  true,
		},
		{
			name: "valid sample query",
			in:   []byte{0x49, 0x00, 0x00, 0x30, 0x39, 0x00, 0x00, 0x00, 0x65},
			want: Message{messageType: 'I', field1: 12345, field2: 101},
		},
	}

	for _, test := range tests {
		got, err := ParseMessage(test.in)

		if test.err {
			if err == nil {
				t.Errorf("%s: expected error but got none", test.name)
			}
			continue
		}

		if err != nil {
			t.Errorf("%s: unexpected error %v", test.name, err)
			continue
		}

		if got != test.want {
			t.Errorf("%s: got %v, want %v", test.name, got, test.want)
		}
	}
}
