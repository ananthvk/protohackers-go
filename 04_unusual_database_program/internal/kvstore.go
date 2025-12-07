package internal

import (
	"fmt"
	"strings"
)

// KVStore is a simple key value store that is implemented using a map. Empty keys are allowed. Note: No locks are used, and hence
// this data structure is unsafe for concurrent access
type KVStore struct {
	mp map[string]string
}

func NewKVStore(version string) *KVStore {
	store := &KVStore{
		mp: make(map[string]string),
	}
	store.mp["version"] = version
	return store
}

// QueryResult holds the result of executing a query. HasValue is set to true if the set Value should be returned to the client, otherwise false
type QueryResult struct {
	Value    string
	HasValue bool
}

func (k *KVStore) ExecuteQuery(query string) QueryResult {
	before, after, found := strings.Cut(query, "=")
	if found {
		// It's an insert / update
		if before == "version" {
			return QueryResult{HasValue: false}
		}
		k.mp[before] = after
		return QueryResult{HasValue: false}

	} else {
		// It's a retrieval
		result := fmt.Sprintf("%s=%s", before, k.mp[before])
		return QueryResult{Value: result, HasValue: true}
	}
}
