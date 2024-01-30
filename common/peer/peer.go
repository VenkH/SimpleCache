package peer

import pb "SimpleCache/common/pb"

// PeerPicker is the interface that must be implemented to locate
// the peer that owns a specific key.
type PeerPicker interface {
	PickPeer(key string) (peer PeerGetter, ok bool) // PickPeer select a getter
}

// PeerGetter is the interface that must be implemented by a peer.
type PeerGetter interface {
	Get(in *pb.Request, out *pb.Response) error
}
