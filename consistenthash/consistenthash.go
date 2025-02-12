package consistenthash

import (
	"fmt"
	"hash/crc32"
	"slices"
)

// A Hash is a function that maps a byte slice to a uint32.
type Hash func([]byte) uint32

// A Ring is a consistent hash ring.
type Ring struct {
	hash    Hash
	weight  int               // the number of virtual nodes for a real node
	nodes   []uint32          // the sorted list of hash values of the nodes
	hashMap map[uint32]string // the mapping of hash values to the names of real nodes
}

func New(weight int, fn Hash) *Ring {
	m := &Ring{
		weight:  weight,
		hash:    fn,
		hashMap: make(map[uint32]string),
	}
	if m.hash == nil {
		m.hash = crc32.ChecksumIEEE
	}
	return m
}

// Add adds nodes to the ring.
// Empty node names are ignored.
func (m *Ring) Add(nodeNames ...string) {
	for _, nodeName := range nodeNames {
		if nodeName == "" {
			continue
		}
		for i := range m.weight {
			hash := m.hash([]byte(fmt.Sprintf("%s_v%d", nodeName, i)))
			m.nodes = append(m.nodes, hash)
			m.hashMap[hash] = nodeName
		}
	}
	slices.Sort(m.nodes)
}

// Get gets the real node for the given key from the ring.
// Returns an empty string if the ring has no nodes.
func (m *Ring) Get(key string) string {
	if len(m.nodes) == 0 {
		return ""
	}

	hash := m.hash([]byte(key))
	idx, _ := slices.BinarySearch(m.nodes, hash)
	// idx%len(m.nodes) == 0 when idx == len(m.nodes) finds the correct node
	return m.hashMap[m.nodes[idx%len(m.nodes)]]
}
