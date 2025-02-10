package gocache

import (
	"fmt"
	"log"
	"net/http"
)

const endpointBasePath = "/gocache"

// A HTTPPool is a pool of HTTP peers.
type HTTPPool struct {
	url      string // this node's base URL
	basePath string // base path for the endpoint
}

func NewHTTPPool(url string) *HTTPPool {
	return &HTTPPool{
		url:      url,
		basePath: endpointBasePath,
	}
}

func (p *HTTPPool) Log(format string, a ...any) {
	log.Printf("[server %s] %s", p.url, fmt.Sprintf(format, a...))
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

		// Set header content type to generic binary data.
		w.Header().Set("Content-Type", "application/octet-stream")
		w.Write(view.ByteSlice())
	})

	// Handle bad requests.
	mux.HandleFunc("/", func(w http.ResponseWriter, r *http.Request) {
		p.Log("bad request: %s", r.URL.Path)
		http.Error(w, "bad request", http.StatusBadRequest)
	})

	return mux
}
