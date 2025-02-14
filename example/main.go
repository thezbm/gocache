package main

import (
	"flag"
	"fmt"
	"log"
	"net/http"
	"time"

	"github.com/thezbm/gocache"
)

// startCacheServer starts a cache server at addr with peer URLs
func startCacheServer(addr string, urls []string, group *gocache.Group) {
	pool := gocache.NewHTTPPool(addr)
	pool.SetPeers(urls...)
	group.RegisterPeers(pool)
	log.Println("gocache server is running at:", addr)
	http.ListenAndServe(addr[7:], pool.GetHTTPHandler())
}

// startAPIServer starts a simple API server for the cache at addr.
func startAPIServer(addr string, group *gocache.Group) {
	http.Handle("/api", http.HandlerFunc(
		func(w http.ResponseWriter, r *http.Request) {
			key := r.URL.Query().Get("key")
			view, err := group.Get(key)
			if err != nil {
				http.Error(w, err.Error(), http.StatusInternalServerError)
				return
			}
			w.Header().Set("Content-Type", "application/octet-stream")
			w.Write(view.ByteSlice())
		}))
	log.Println("gocache API server is running at:", addr)
	http.ListenAndServe(addr[7:], nil)
}

func main() {
	var port int
	var api bool
	flag.IntVar(&port, "port", 8000, "set the port for the cache server")
	flag.BoolVar(&api, "api", false, "start the API server")
	flag.Parse()

	apiURL := "http://0.0.0.0:9999"
	peerURLs := []string{
		"http://localhost:8001",
		"http://localhost:8002",
		"http://localhost:8003",
	}

	group := gocache.NewGroup("example", 1<<10, gocache.GetterFunc(
		func(key string) ([]byte, error) {
			time.Sleep(time.Second * 1) // simulate a slow database call
			return []byte(key), nil
		}))
	if api {
		go startAPIServer(apiURL, group)
	}
	startCacheServer(fmt.Sprintf("http://localhost:%d", port), peerURLs, group)
}
