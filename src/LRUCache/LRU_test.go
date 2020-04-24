package cache

import (
	"fmt"
	"testing"
)

func TestLRU(t *testing.T) {
	cache := NewCache(2)
	cache.Put("1", "one")
	fmt.Println(cache.Get("1"))
	cache.Put("2", "two")
	fmt.Println(cache.Get("1"))
	cache.Put("3", "three")
	fmt.Println(cache.Get("2"))
	fmt.Println(cache.Get("3"))
	fmt.Println(cache.Get("3"))
	fmt.Println(cache.Get("1"))
	cache.Put("2", "two")
	fmt.Println(cache.Get("3"))
	fmt.Println(cache.Get("1"))
}
