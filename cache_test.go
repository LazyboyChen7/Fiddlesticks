package cache

import "testing"

func BenchmarkPut(t *testing.B) {
	c := NewCache(5)
	for i := 0; i < t.N; i++ {
		c.Put(i, i)
	}
}

func BenchmarkGet(t *testing.B) {
	c := NewCache(5)
	for i := 0; i < t.N; i++ {
		c.Put(i, i)
	}
	for i := 0; i < t.N; i++ {
		c.Get(i)
	}
}
