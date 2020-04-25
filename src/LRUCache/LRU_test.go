package cache

import (
	"math/rand"
	"strconv"
	"testing"
)

func TestLRU(t *testing.T) {
	cache := NewCache(2)
	cache.Put("1", "one")
	t.Log(cache.Get("1"))
	cache.Put("2", "two")
	t.Log(cache.Get("1"))
	cache.Put("3", "three")
	t.Log(cache.Get("2"))
	t.Log(cache.Get("3"))
	t.Log(cache.Get("3"))
	t.Log(cache.Get("1"))
	cache.Put("2", "two")
	t.Log(cache.Get("3"))
	t.Log(cache.Get("1"))
}

func Benchmark_Put(t *testing.B) {
	cache := NewCache(600)
	for i := 0; i < t.N; i++ {
		cache.Put(strconv.Itoa(i), i)
	}
}

func Benchmark_PutGet(t *testing.B) {
	cache := NewCache(600)
	for i := 0; i < t.N; i++ {
		cache.Put(strconv.Itoa(i), i)
		cache.Get(strconv.Itoa(rand.Intn(2000)))
	}
}
