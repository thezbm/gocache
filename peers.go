package gocache

import pb "github.com/thezbm/gocache/gocachepb"

// A PeerPicker is able to pick a peer based on the key.
type PeerPicker interface {
	PickPeer(key string) (Peer, bool)
}

// A Peer is able to get data for the given group and key.
type Peer interface {
	Get(in *pb.Request, out *pb.Response) error
}
