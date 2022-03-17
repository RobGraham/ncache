package ncache

import (
	"math/rand"
	"sync"
	"testing"
	"time"
)

func TestNew(t *testing.T) {

	t.Run("Creates a non-nil instance of `Cache`", func(t *testing.T) {
		if c, err := New(&Config{}); c == nil || err != nil {
			t.Errorf("cache was nil or error occurred: \n%s ", err)
		}
	})

	t.Run("Passing nil `config` should error", func(t *testing.T) {
		if c, err := New(nil); c != nil && err == nil {
			t.Fail()
		}
	})

	t.Run("Should execute eviction callback", func(t *testing.T) {
		key := "123"
		msgChan := make(chan string)

		c, _ := New(&Config{
			Evict: 20 * time.Millisecond,
			OnEvict: func(key, value interface{}) {
				msgChan <- key.(string)
			},
		})

		c.Set(key, "abc", 10*time.Millisecond)

		o := <-msgChan
		if o != key {
			t.Fail()
		}
	})
}

func TestAdd(t *testing.T) {
	c, _ := New(&Config{})

	t.Run("Should store a new entry", func(t *testing.T) {
		if err := c.Add("a", "b", 0); err != nil {
			t.Error(err)
		}
		if _, ok := c.Get("a"); !ok {
			t.Error("Expected entry to exist")
		}
	})

	t.Run("Should error if key already exists", func(t *testing.T) {
		if err := c.Add("a", "b", 0); err == nil {
			t.Fail()
		}
	})
}

func TestSet(t *testing.T) {
	c, _ := New(&Config{})

	t.Run("Should store a new entry", func(t *testing.T) {
		c.Set("a", "b", 0)

		if _, ok := c.Get("a"); !ok {
			t.Error("Expected entry to exist")
		}
	})

	t.Run("Should override existing key", func(t *testing.T) {
		c.Set("a", "c", 0)

		if v, ok := c.Get("a"); !ok || v != "c" {
			t.Error("Expected entry to exist with new value")
		}
	})
}

func TestGet(t *testing.T) {
	c, _ := New(&Config{})

	t.Run("Should return an entry", func(t *testing.T) {
		c.Set("a", "b", 0)

		if _, ok := c.Get("a"); !ok {
			t.Error("Expected entry to exist")
		}
	})

	t.Run("Should not return an expired entry", func(t *testing.T) {
		c.Set("1", "2", 10*time.Millisecond)
		time.Sleep(11 * time.Millisecond)

		if _, ok := c.Get("1"); ok {
			t.Fail()
		}
	})

	t.Run("Should return a non-expired entry", func(t *testing.T) {
		c.Set("foo", "bar", 20*time.Millisecond)
		time.Sleep(10 * time.Millisecond)

		if _, ok := c.Get("foo"); !ok {
			t.Fail()
		}
	})
}

func TestDelete(t *testing.T) {
	c, _ := New(&Config{})

	t.Run("Should delete an entry", func(t *testing.T) {
		c.Set("a", "b", 0)
		c.Delete("a")

		if _, ok := c.Get("a"); ok {
			t.Error("Found entry")
		}
	})
}

func TestFlush(t *testing.T) {
	c, _ := New(&Config{})

	t.Run("Should delete an entry", func(t *testing.T) {
		c.Set("a", "b", 0)
		c.Flush()

		if _, ok := c.Get("a"); ok {
			t.Error("Found entry")
		}
	})
}

func TestCache_evictor(t *testing.T) {

	t.Run("Should execute eviction when evict time is provided", func(t *testing.T) {
		key := "123"

		c, _ := New(&Config{
			Evict: 10 * time.Millisecond,
		})

		c.Set(key, "abc", 20*time.Millisecond)

		time.Sleep(30 * time.Millisecond)

		// Bypass `cache.Get` which doesn't return entry if TTL has expired
		if _, ok := c.entries.Load(key); ok {
			t.Error("Entry should not exist")
		}
	})

	t.Run("Should not execute eviction when evict is empty", func(t *testing.T) {
		key := "123"

		c, _ := New(&Config{})

		c.Set(key, "abc", 10*time.Millisecond)
		time.Sleep(15 * time.Millisecond)

		// Bypass `cache.Get` which doesn't return entry if TTL has expired
		if _, ok := c.entries.Load(key); !ok {
			t.Error("Entry should still exist")
		}
	})
}

func TestConcurrency(t *testing.T) {
	c, _ := New(&Config{})
	a := [10000]int{}
	rand.Seed(time.Now().UnixNano())

	var wg sync.WaitGroup

	wg.Add(2)
	go func() {
		defer wg.Done()
		for range a {
			c.Set("k", rand.Intn(1000-10+1)+10, 0)
		}
	}()

	go func() {
		defer wg.Done()
		for range a {
			c.Get("k")
		}
	}()
	wg.Wait()
}
