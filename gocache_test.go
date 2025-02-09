package gocache

import (
	"fmt"
	"reflect"
	"testing"
)

func TestGetter(t *testing.T) {
	f := GetterFunc(func(key string) ([]byte, error) {
		return []byte(key), nil
	})

	expected := []byte("key")
	if v, _ := f.Get("key"); !reflect.DeepEqual(expected, v) {
		t.Fatalf("cache getter callback failed (expected: %v, got: %v)", expected, v)
	}
}

func TestGet(t *testing.T) {
	// db is a mock database
	var db = map[string]string{
		"Alice":   "123",
		"Bob":     "456",
		"Charlie": "789",
	}
	loadCounts := map[string]int{}
	g := NewGroup("numbers", 0, GetterFunc(
		func(key string) ([]byte, error) {
			if v, ok := db[key]; ok {
				if _, ok := loadCounts[key]; !ok {
					loadCounts[key] = 0
				}
				loadCounts[key]++
				return []byte(v), nil
			}
			return nil, fmt.Errorf("key=%s does not exist", key)
		}))

	for k, v := range db {
		if view, err := g.Get(k); err != nil || view.String() != v {
			t.Fatalf("cache Get failed with key=%s (expected: %s, got: %s)", k, v, view.String())
		}
		if _, err := g.Get(k); err != nil || loadCounts[k] > 1 {
			t.Fatalf("cache Get failed to hit with key=%s", k)
		}
	}
	if view, err := g.Get("Daniel"); err == nil {
		t.Fatalf("cache Get failed with key=Daniel (expected an error, got: %v)", view)
	}
}

func TestGetGroup(t *testing.T) {
	groupName := "testGroup"
	NewGroup(groupName, 2<<10, GetterFunc(
		func(key string) (_ []byte, _ error) { return }))

	if group := GetGroup(groupName); group == nil || group.name != groupName {
		t.Fatal("GetGroup failed")
	}
	if group := GetGroup("testGroup2"); group != nil {
		t.Fatalf("GetGroup failed with key=testGroup2 (expected an error, got: %s)", group.name)
	}
}
