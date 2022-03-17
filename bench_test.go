package ncache

import "testing"

func BenchmarkSet(b *testing.B) {
	c, _ := New(&Config{})

	b.RunParallel(func(pb *testing.PB) {
		for pb.Next() {
			c.Set("k", "v", 0)
		}
	})
}

func BenchmarkGet(b *testing.B) {
	c, _ := New(&Config{})
	c.Set("k", "v", 0)

	b.RunParallel(func(pb *testing.PB) {
		for pb.Next() {
			c.Get("k")
		}
	})
}
