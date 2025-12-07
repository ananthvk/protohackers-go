package internal

import "testing"

func TestKVStore(t *testing.T) {
	// All the queries will be executed one after the other with the same store
	tests := []struct {
		query string
		want  QueryResult
	}{
		// Simple insertions
		{"foo=bar", QueryResult{HasValue: false}},
		{"foo=bar", QueryResult{HasValue: false}},

		// Simple retrieval
		{"foo", QueryResult{Value: "foo=bar", HasValue: true}},

		// Update
		{"foo=baz", QueryResult{HasValue: false}},
		{"foo", QueryResult{Value: "foo=baz", HasValue: true}},

		// Non existent key
		{"not_exist", QueryResult{Value: "not_exist=", HasValue: true}},

		// Empty keys
		{"", QueryResult{Value: "=", HasValue: true}},
		{"=some value here", QueryResult{HasValue: false}},
		{"", QueryResult{Value: "=some value here", HasValue: true}},
		{"=This has been updated", QueryResult{HasValue: false}},
		{"", QueryResult{Value: "=This has been updated", HasValue: true}},
		{"=====", QueryResult{HasValue: false}},
		{"", QueryResult{Value: "=====", HasValue: true}},

		// Empty values
		{"foobar=", QueryResult{HasValue: false}},
		{"foobar", QueryResult{Value: "foobar=", HasValue: true}},

		// Empty keys & values
		{"=", QueryResult{HasValue: false}},
		{"", QueryResult{Value: "=", HasValue: true}},

		// Queries with multiple equals
		{"foo=bar=baz", QueryResult{HasValue: false}},
		{"key2=another=key", QueryResult{HasValue: false}},
		{"foo1===", QueryResult{HasValue: false}},
		{"foo", QueryResult{Value: "foo=bar=baz", HasValue: true}},
		{"key2", QueryResult{Value: "key2=another=key", HasValue: true}},
		{"foo1", QueryResult{Value: "foo1===", HasValue: true}},

		// Trailing spaces
		{"   foo=some", QueryResult{HasValue: false}},
		{"    4 spaces    =    value    ", QueryResult{HasValue: false}},
		{"   foo", QueryResult{Value: "   foo=some", HasValue: true}},
		{"    4 spaces    ", QueryResult{Value: "    4 spaces    =    value    ", HasValue: true}},

		// Test that version is immutable
		{"version", QueryResult{Value: "version=1.0.0-test", HasValue: true}},
		{"version=", QueryResult{HasValue: false}},
		{"version", QueryResult{Value: "version=1.0.0-test", HasValue: true}},
		{"version=some-fake-version", QueryResult{HasValue: false}},
		{"version", QueryResult{Value: "version=1.0.0-test", HasValue: true}},

		// Test that keys with 'version' in them are valid
		{" version=some-fake-version", QueryResult{HasValue: false}},
		{" version", QueryResult{Value: " version=some-fake-version", HasValue: true}},
	}

	store := NewKVStore("1.0.0-test")
	for _, test := range tests {
		result := store.ExecuteQuery(test.query)
		if result != test.want {
			t.Errorf("got %v, want %v for query: %v", result, test.want, test.query)
		}
	}
}
