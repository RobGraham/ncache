# Introduction

ncache is a simple Golang in-memory caching solution utilizing sync.Map which is best suited for read heavy cache usage.

# Getting started
```go
cache, err := ncache.New(&ncache.Config{
    Evict: 10 * time.Minutes, // optional
    OnEvict: func(key, value interface{}) { // optional
        // ... do stuff
    },
})

cache.Set("abc", "some data", 30 * time.Seconds) // Add to cache with TTL of 30 seconds

if v, ok := cache.Get("abc"); ok {
    // Cache found
}
```

# Benchmark
The decision to use sync.Map instead of regular map with Mutex was the significantly improved read times.

Running on a single cpu core (i7-9750H) the results are:

### sync.Map
| Test | No. Ops | Speed | Bytes/op | Allocations |
| ---- | ---- | ---- | ---- | ---- |
| BenchmarkSet | 9887005 | 114.3 ns/op | 40 B/op | 2 allocs/op
| BenchmarkGet | 55380897 | 23.23 ns/op | 0 B/op | 0 allocs/op


### Map with mutex
| Test | No. Ops | Speed | Bytes/op | Allocations |
| ---- | ---- | ---- | ---- | ---- |
| BenchmarkSet | 15337158 | 65.39 ns/op | 16 B/op | 1 allocs/op |
| BenchmarkGet | 31743174 | 32.49 ns/op | 0 B/op | 0 allocs/op |


However, the results of sync.Map are more pronounced if using a multi cpu system.

Using all cores (6) of an i7-9750H you'll notice that read is almost 5x faster while write suffers a little.

| Test | No. Ops | Speed | Bytes/op | Allocations |
| ---- | ---- | ---- | ---- | ---- |
| BenchmarkSet | 6128324 | 187.5 ns/op | 40 B/op | 2 allocs/op
| BenchmarkGet | 234110263 | 5.438 ns/op | 0 B/op | 0 allocs/op
