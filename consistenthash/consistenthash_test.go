package consistenthash

import (
	"slices"
	"testing"
)

func TestRing(t *testing.T) {
	// Create the ring:
	// Nodes:   N1_v0  N2_v1  N3_v0  N1_v2  N2_v2  N1_v1  N2_v0  N3_v2  N3_v1
	// Hashes:    10     20     25     30     38     49     70     82    90
	// Keys:    k2   k1             k4        k5                k6         k3
	// Hashes:  00   11             28        38                75         93
	nodeHashes := []uint32{10, 49, 30, 70, 20, 38, 25, 90, 82}
	keyHashes := []uint32{11, 0, 93, 28, 38, 75}
	hashes := append(nodeHashes, keyHashes...)
	counter := 0
	r := New(3, func(key []byte) uint32 {
		hash := hashes[counter]
		counter += 1
		return hash
	})
	r.Add("node1", "node2", "node3")

	testCases := []struct {
		key      string
		expected string
	}{
		{"key1", "node2"},
		{"key2", "node1"},
		{"key3", "node1"},
		{"key4", "node1"},
		{"key5", "node2"},
		{"key6", "node3"},
	}
	for _, testCase := range testCases {
		if node := r.Get(testCase.key); node != testCase.expected {
			t.Fatalf("consistenthash ring failed with key=%s (expected: %s, got %s)",
				testCase.key, testCase.expected, node)
		}
	}

	// Get rid of node2.
	r.nodes = slices.DeleteFunc(r.nodes, func(node uint32) bool {
		return node == 20 || node == 38 || node == 70
	})
	hashes = keyHashes
	counter = 0

	testCases = []struct {
		key      string
		expected string
	}{
		{"key1", "node3"},
		{"key2", "node1"},
		{"key3", "node1"},
		{"key4", "node1"},
		{"key5", "node1"},
		{"key6", "node3"},
	}
	for _, testCase := range testCases {
		if node := r.Get(testCase.key); node != testCase.expected {
			t.Fatalf("consistenthash ring failed with key=%s (expected: %s, got %s)",
				testCase.key, testCase.expected, node)
		}
	}
}
