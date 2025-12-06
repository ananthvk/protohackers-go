package internal

import (
	"testing"
)

func TestParseRequest(t *testing.T) {
	table := []struct {
		in      string
		want    Request
		wantErr bool
	}{
		// Test empty requests
		{"", Request{}, true},
		{"\n\n", Request{}, true},
		{"    ", Request{}, true},

		// Test malformed requests
		{"{", Request{}, true},
		{"[]}", Request{}, true},
		{"{}", Request{}, true},

		// Extra characters after JSON object
		{`{"method": "isPrime", "number": 2}a`, Request{}, true},
		{`{"method": "isPrime", "number": 2}{}`, Request{}, true},
		{`{"method": "isPrime", "number": 8}{"another": "json"}`, Request{}, true},
		{`{"method": "isPrime", "number": 9}[1,2,3]`, Request{}, true},

		// Test missing fields
		{"{}", Request{}, true},
		{`{"method": "isPrime"}`, Request{}, true},
		{`{"number": 32}`, Request{}, true},

		// Test invalid method type
		{`{"method": "isprime", "number": 32}`, Request{}, true},
		{`{"method": "", "number": 32}`, Request{}, true},
		{`{"method": "isNotPrime", "number": 32}`, Request{}, true},

		// Test invalid number type
		{`{"method": "isPrime", "number": true}`, Request{}, true},
		{`{"method": "isPrime", "number": {}}`, Request{}, true},
		{`{"method": "isPrime", "number": "a string"}`, Request{}, true},
		{`{"method": "isPrime", "number": "32"}`, Request{}, true},
		{`{"method": "isPrime", "number": "32.15"}`, Request{}, true},

		// Test valid inputs
		{`{"method": "isPrime", "number": 0}`, Request{0}, false},
		{`{"method": "isPrime", "number": -1}`, Request{-1}, false},
		{`{"method": "isPrime", "number": 2}`, Request{2}, false},
		{`{"method": "isPrime", "number": 123456}`, Request{123456}, false},
		{`{"method": "isPrime", "number": 3.1415}`, Request{3}, false},
		{`{"method": "isPrime", "number": 3.999}`, Request{3}, false},
		{`{"method": "isPrime", "number": 8.9999999}`, Request{8}, false},
		{`{"method": "isPrime", "number": 0.01}`, Request{0}, false},
		{`{"method": "isPrime", "number": -0.01}`, Request{0}, false},

		// Test extraneous fields
		{`{"method": "isPrime", "number": 23, "other_input":[1,2,3,4]}`, Request{23}, false},
	}
	for _, row := range table {
		got, err := ParseRequest([]byte(row.in))
		if row.wantErr {
			if err == nil {
				t.Errorf("ParseRequest(%q) expected error", row.in)
			}
			continue
		}
		if err != nil {
			t.Errorf("ParseRequest(%q) produced unexpected error: %v", row.in, err)
			continue
		}
		if got != row.want {
			t.Errorf("ParseRequest(%q) : got %v want %v", row.in, got, row.want)
		}
	}
}
