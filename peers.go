package gocache

// A PeerPicker is able to pick a peer based on the key.
type PeerPicker interface {
	PickPeer(key string) (Peer, bool)
}

// A Peer is able to get data for the given group and key.
type Peer interface {
	Get(group string, key string) ([]byte, error)
}
