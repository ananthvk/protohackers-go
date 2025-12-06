package internal

import "testing"

func TestIsPrime(t *testing.T) {
	cases := []struct {
		in  int64
		out bool
	}{
		{-1, false},
		{0, false},
		{1, false},
		{2, true},
		{3, true},
		{4, false},
		{5, true},
		{6, false},
		{7, true},
		{8, false},
		{9, false},
		{10, false},
		{11, true},
		{4535189, true},
		{7474967, true},
		{9737333, true},
		{718064159, true},
		{997525853, true},
		{997525855, false},
		{33900388113763, false},
		{33900388113767, false},
	}
	for _, c := range cases {
		got := IsPrime(c.in)
		if got != c.out {
			t.Errorf("IsPrime(%d) = %v, want %v", c.in, got, c.out)
		}
	}
}
