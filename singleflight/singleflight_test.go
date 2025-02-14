package singleflight

import (
	"sync"
	"testing"
	"time"
)

func TestSingleflight(t *testing.T) {
	keys := []string{"a", "a", "a", "b", "c", "c"}
	sg := Group{}
	wg := new(sync.WaitGroup)
	for _, key := range keys {
		wg.Add(1)
		go func() {
			sg.Do(key, func() (any, error) {
				return dbCall(key)
			})
			wg.Done()
		}()
	}
	wg.Wait()
	if dbAccess != 3 {
		t.Errorf("singleflight failed (expected: %d, got: %d)", 3, dbAccess)
	}
}

var (
	dbAccess int = 0
	mu       sync.Mutex
)

func dbCall(key string) (string, error) {
	mu.Lock()
	dbAccess++
	mu.Unlock()
	time.Sleep(100 * time.Millisecond)
	return key, nil
}
