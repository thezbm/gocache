package lru

import (
	"reflect"
	"testing"
)

type value string

func (v value) Len() int {
	return len(v)
}

func TestGet(t *testing.T) {
	lru := New(int64(0), nil)
	lru.Set("k1", value("v1"))
	if v, ok := lru.Get("k1"); !ok || string(v.(value)) != "v1" {
		t.Fatalf("cache hit k1=v1 failed")
	}
	if _, ok := lru.Get("k2"); ok {
		t.Fatalf("cache miss k2 failed")
	}
}

func TestSet(t *testing.T) {
	lru := New(int64(0), nil)
	lru.Set("k1", value("v1"))
	if lru.size != int64(len("k1"+"v1")) {
		t.Fatalf("cache set failed (expected: %v, got: %v)", len("k1"+"v1"), lru.size)
	}
	lru.Set("k1", value("value1"))
	if lru.size != int64(len("k1"+"value1")) {
		t.Fatalf("cache set failed (expected: %v, got: %v)", len("k1"+"value1"), lru.size)
	}
}

func TestEvict(t *testing.T) {
	k1, k2, k3 := "k1", "k2", "k3"
	v1, v2, v3 := "v1", "v2", "v3"
	capacity := int64(len(k1 + k2 + v1 + v2))
	lru := New(capacity, nil)
	lru.Set(k1, value(v1))
	lru.Set(k2, value(v2))
	lru.Set(k3, value(v3))
	if _, ok := lru.Get("k1"); ok || lru.Len() != 2 {
		t.Fatalf("cache evict k1 failed")
	}
}

func TestOnEvict(t *testing.T) {
	keys := []string{}
	onEvict := func(key string, value Value) {
		keys = append(keys, key)
	}
	k1, k2, k3 := "k1", "k2", "k_3"
	v1, v2, v3 := "v1", "v2", "v_3"
	capacity := int64(len(k1 + k2 + v1 + v2))
	lru := New(capacity, onEvict)
	lru.Set(k1, value(v1))
	lru.Set(k2, value(v2))
	lru.Set(k3, value(v3))
	expected := []string{"k1", "k2"}
	if !reflect.DeepEqual(expected, keys) {
		t.Fatalf("cache onEvict callback failed (expected: %v, got: %v)", expected, keys)
	}
}
