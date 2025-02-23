package gocache

import (
	"fmt"
	"io"
	"log"
	"net/http"
	"net/url"
	"sync"

	"github.com/thezbm/gocache/consistenthash"
	pb "github.com/thezbm/gocache/gocachepb"

	"google.golang.org/protobuf/proto"
)

const (
	endpointBasePath = "/gocache"
	defaultWeight    = 50 // the default weight for the consistent hash ring
)

// A HTTPPool is a pool of HTTP peers.
// It implements the PeerPicker interface.
type HTTPPool struct {
	selfURL   string               // this node's base URL
	basePath  string               // base path for the endpoint
	mu        sync.Mutex           // protects ring and httpPeers
	ring      *consistenthash.Ring // the consistent hash ring
	httpPeers map[string]*httpPeer // maps peer URLs to httpPeer instances
}

func NewHTTPPool(selfURL string) *HTTPPool {
	return &HTTPPool{
		selfURL:  selfURL,
		basePath: endpointBasePath,
	}
}

func (p *HTTPPool) Log(format string, a ...any) {
	log.Printf("[server %s] %s", p.selfURL, fmt.Sprintf(format, a...))
}

// GetHTTPHandler returns the HTTP handler for the pool.
func (p *HTTPPool) GetHTTPHandler() http.Handler {
	mux := http.NewServeMux()

	// Handle GET /<basePath>/<groupname>/<key>.
	pattern := fmt.Sprintf("GET %s/{group}/{key}", p.basePath)
	mux.HandleFunc(pattern, func(w http.ResponseWriter, r *http.Request) {
		p.Log("%s %s", r.Method, r.URL.Path)
		groupName, key := r.PathValue("group"), r.PathValue("key")

		group := GetGroup(groupName)
		if group == nil {
			http.Error(w, "group not found: "+groupName, http.StatusNotFound)
			return
		}

		view, err := group.Get(key)
		if err != nil {
			http.Error(w, err.Error(), http.StatusInternalServerError)
			return
		}

		body, err := proto.Marshal(&pb.Response{Value: view.ByteSlice()})
		if err != nil {
			http.Error(w, err.Error(), http.StatusInternalServerError)
			return
		}

		// Set header content type to generic binary data.
		w.Header().Set("Content-Type", "application/octet-stream")
		w.Write(body)
	})

	// Handle bad requests.
	mux.HandleFunc("/", func(w http.ResponseWriter, r *http.Request) {
		p.Log("bad request: %s", r.URL.Path)
		http.Error(w, "bad request", http.StatusBadRequest)
	})

	return mux
}

// SetPeers sets the peers for the pool with their base URLs.
func (p *HTTPPool) SetPeers(peers ...string) {
	p.mu.Lock()
	defer p.mu.Unlock()

	p.ring = consistenthash.New(defaultWeight, nil)
	p.ring.Add(peers...)

	p.httpPeers = make(map[string]*httpPeer, len(peers))
	for _, peer := range peers {
		p.httpPeers[peer] = &httpPeer{baseURL: peer + p.basePath}
	}
}

// It returns the HTTP peer for the given key.
// If the ring is empty or the peer is this node itself, it returns nil and false.
func (p *HTTPPool) PickPeer(key string) (Peer, bool) {
	p.mu.Lock()
	defer p.mu.Unlock()

	if peer := p.ring.Get(key); peer != "" && peer != p.selfURL {
		p.Log("pick peer %s", peer)
		return p.httpPeers[peer], true
	}
	return nil, false
}

// An httpPeer implements the Peer interface.
// It is the HTTP client for accessing the remote peer.
type httpPeer struct {
	baseURL string
}

// Get sends a GET request to the remote peer for the given group and key in the Protocol Buffer request.
func (h *httpPeer) Get(in *pb.Request, out *pb.Response) error {
	url := fmt.Sprintf("%s/%s/%s",
		h.baseURL, url.QueryEscape(in.GetGroup()), url.QueryEscape(in.GetKey()))
	resp, err := http.Get(url)
	if err != nil {
		return err
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		return fmt.Errorf("endpoint %s returned: %s", h.baseURL, resp.Status)
	}

	bytes, err := io.ReadAll(resp.Body)
	if err != nil {
		return fmt.Errorf("reading response body: %v", err)
	}

	if err = proto.Unmarshal(bytes, out); err != nil {
		return fmt.Errorf("decoding response body: %v", err)
	}
	return nil
}
