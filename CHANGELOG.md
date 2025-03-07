# Changelog

## [v1.0.0] (2025-03-7)

### 🎉 Initial Release

GoCQ (Go Concurrent Queue) is a high-performance concurrent queue system for Go, featuring both FIFO and priority queue implementations.

### ✨ Features

- Generic type support for both data and results
- Configurable worker pool with controlled concurrency
- Two queue implementations:
  - Standard FIFO Queue with O(1) operations
  - Priority Queue with O(log n) operations
- Non-blocking job submission
- Thread-safe operations
- Pause/Resume functionality
- Clean shutdown mechanisms
- Comprehensive test coverage

### 🔧 Technical Details

- Minimum Go version: 1.24+
- Standard Queue: Based on `container/list`
- Priority Queue: Based on `container/heap`
- Thread safety using sync primitives
- Memory-efficient channel management

### 🚀 Performance

Benchmark results on Intel i7-13700:

```bash
goos: linux
goarch: amd64
pkg: github.com/fahimfaisaal/gocq
cpu: 13th Gen Intel(R) Core(TM) i7-13700
BenchmarkPriorityQueue_Operations/Add-24                 1378249              1278 ns/op
BenchmarkPriorityQueue_Operations/AddAll-24               795332              1712 ns/op
BenchmarkQueue_Operations/Add-24                         1000000              1300 ns/op
BenchmarkQueue_Operations/AddAll-24                      1000000              1822 ns/op
```

### 📝 License

MIT License - See LICENSE file for details
