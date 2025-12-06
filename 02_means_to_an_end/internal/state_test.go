package internal

import "testing"

// Execute a list of operations, the last op should be a query to check the resultant value
type op struct {
	kind   string
	field1 int32
	field2 int32
}

func TestState(t *testing.T) {
	tests := []struct {
		name string
		ops  []op
		want int32
	}{
		{
			name: "empty state, no ops",
			ops:  []op{},
			want: 0,
		},
		{
			name: "query with no inserts",
			ops: []op{
				{"Q", 0, 100},
			},
			want: 0,
		},
		{
			name: "single insert + query hit",
			ops: []op{
				{"I", 400, 9999},
				{"Q", 300, 600},
			},
			want: 9999,
		},
		{
			name: "single insert + query miss below",
			ops: []op{
				{"I", 400, 9999},
				{"Q", 0, 399},
			},
			want: 0,
		},
		{
			name: "single insert + query miss above",
			ops: []op{
				{"I", 400, 9999},
				{"Q", 401, 999},
			},
			want: 0,
		},

		// -------------------------------------------------------
		// Out-of-order inserts
		// -------------------------------------------------------
		{
			name: "out-of-order inserts",
			ops: []op{
				{"I", 5, 50},
				{"I", 1, 10},
				{"I", 3, 30},
				{"Q", 1, 5},
			},
			want: 30,
		},

		// -------------------------------------------------------
		// Multiple inserts + ranges
		// -------------------------------------------------------
		{
			name: "multiple inserts average all",
			ops: []op{
				{"I", 1, 10},
				{"I", 2, 20},
				{"I", 3, 30},
				{"I", 4, 40},
				{"I", 5, 50},
				{"Q", 1, 5},
			},
			want: 30,
		},
		{
			name: "multiple inserts average subset",
			ops: []op{
				{"I", 1, 10},
				{"I", 2, 20},
				{"I", 3, 30},
				{"I", 4, 40},
				{"I", 5, 50},
				{"Q", 2, 5},
			},
			want: 35,
		},
		{
			name: "query exact match single element",
			ops: []op{
				{"I", 3, 30},
				{"Q", 3, 3},
			},
			want: 30,
		},
		{
			name: "query lower bound inclusive",
			ops: []op{
				{"I", 3, 30},
				{"I", 5, 50},
				{"Q", 3, 4},
			},
			want: 30,
		},
		{
			name: "query upper bound inclusive",
			ops: []op{
				{"I", 3, 30},
				{"I", 5, 50},
				{"Q", 4, 5},
			},
			want: 50,
		},

		// -------------------------------------------------------
		// Edge: min > max
		// -------------------------------------------------------
		{
			name: "invalid range min > max",
			ops: []op{
				{"I", 10, 111},
				{"Q", 20, 10},
			},
			want: 0,
		},

		// -------------------------------------------------------
		// Negative timestamps and prices
		// -------------------------------------------------------
		{
			name: "negative timestamps",
			ops: []op{
				{"I", -10, 100},
				{"I", -5, 300},
				{"Q", -20, 0},
			},
			want: 200,
		},
		{
			name: "negative prices",
			ops: []op{
				{"I", 1, -100},
				{"I", 2, 300},
				{"Q", 1, 2},
			},
			want: 100,
		},

		// -------------------------------------------------------
		// Big range
		// -------------------------------------------------------
		{
			name: "big range covers all",
			ops: []op{
				{"I", 1, 10},
				{"I", 1000, 20},
				{"Q", -9999, 999999},
			},
			want: 15,
		},

		// -------------------------------------------------------
		// Overlapping queries
		// -------------------------------------------------------
		{
			name: "overlapping ranges",
			ops: []op{
				{"I", 1, 10},
				{"I", 5, 50},
				{"Q", 1, 5}, // = 30
				{"I", 3, 30},
				{"Q", 1, 5}, // = 30
			},
			want: 30,
		},

		// -------------------------------------------------------
		// Repeat query immutability
		// -------------------------------------------------------
		{
			name: "repeated queries same result",
			ops: []op{
				{"I", 1, 10},
				{"I", 2, 20},
				{"Q", 1, 5},
				{"Q", 1, 5},
			},
			want: 15,
		},
	}

	for _, test := range tests {
		state := State{}
		got := int32(0)
		for _, op := range test.ops {
			switch op.kind {
			case "I":
				state.Insert(op.field1, op.field2)
			case "Q":
				got = state.QueryAverage(op.field1, op.field2)
			default:
				panic("Invalid operation type")
			}
		}
		if got != test.want {
			t.Errorf("test %s: got %v, want %v", test.name, got, test.want)
		}
	}
}
