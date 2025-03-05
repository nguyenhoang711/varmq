# GoCQ - Go Concurrent Queue

GoCQ is a high-performance, generic concurrent queue implementation in Go that supports both regular FIFO queues and priority queues with configurable concurrency levels.

## Features

- Generic type support for both data and results
- Configurable concurrency levels
- Priority queue support
- Non-blocking operations
- Channel-based communication
- Simple and intuitive API
- Thread-safe operations

## Installation

```bash
go get github.com/fahimfaisaal/gocq
```

## Usage

### Regular Concurrent Queue

```go
// Create a new queue with 4 concurrent workers
queue := gocq.New(4, func(data int) int {
    // Worker function that processes the data
    return data * 2
})
defer queue.Close()

// Add a single item
result := <-queue.Add(5) // result will be 10

// Add multiple items
numbers := []int{1, 2, 3, 4, 5}
results := queue.AddAll(numbers...)
for result := range results {
    // Process results as they complete
    fmt.Println(result)
}
```

### Priority Queue

```go
// Create a priority queue with 4 concurrent workers
pq := gocq.NewPQ(4, func(data int) int {
    return data * 2
})
defer pq.Close()

// Add items with different priorities (lower number = higher priority)
result1 := <-pq.Add(5, 1)  // priority 1
result2 := <-pq.Add(10, 2) // priority 2

// Add multiple items with the same priority
numbers := []int{1, 2, 3, 4, 5}
results := pq.AddAll(1, numbers...) // all with priority 1
for result := range results {
    fmt.Println(result)
}
```

## API Reference

### Initialization

- `Init() *concurrentQueue[T, R]`
  - Initializes the queue by starting worker goroutines
  - Called automatically by `New()` and `NewPQ()`
  - Returns the initialized queue
  - Time complexity: O(n) where n is the concurrency level

### Regular Queue

- `New[T, R any](concurrency uint, worker func(T) R) *concurrentQueue[T, R]`

  - Creates a new concurrent queue with specified concurrency level
  - Generic types T (input) and R (output)
  - Automatically calls `Init()` to start worker goroutines

- `Add(data T) <-chan R`

  - Adds a single item to the queue
  - Returns a channel for the result

- `AddAll(data ...T) <-chan R`
  - Adds multiple items to the queue
  - Returns a channel that receives all results

### Priority Queue

- `NewPQ[T, R any](concurrency uint, worker func(T) R) *concurrentPriorityQueue[T, R]`

  - Creates a new concurrent priority queue
  - Generic types T (input) and R (output)
  - Automatically calls `Init()` to start worker goroutines

- `Add(data T, priority int) <-chan R`

  - Adds an item with specified priority
  - Lower priority number = higher priority

- `AddAll(priority int, data ...T) <-chan R`
  - Adds multiple items with the same priority

### Common Methods

- `WaitUntilFinished()`

  - Waits for all pending jobs to complete

- `Purge()`

  - Removes all pending jobs from the queue
  - Closes response channels for pending jobs
  - Does not affect currently processing jobs
  - Useful for clearing the queue without waiting for completion
  - Thread-safe operation

- `Close()`

  - Closes the queue and cleans up resources
  - Calls `Purge()` internally
  - Waits for ongoing processes to complete
  - Closes all worker channels
  - Resets internal channels and states

- `WaitAndClose()`

  - Waits for completion and closes the queue
  - Combination of `WaitUntilFinished()` and `Close()`

- `PendingCount() int`

  - Returns the number of pending jobs

- `CurrentProcessingCount() uint`
  - Returns the number of jobs currently being processed

### Usage Example with Purge

```go
queue := gocq.New(4, func(data int) int {
    return data * 2
})
defer queue.Close()

// Add some items
for i := 0; i < 100; i++ {
    queue.Add(i)
}

// Clear all pending jobs without waiting
queue.Purge()

// Add new items after purge
queue.Add(1)
```

## Internal Implementation Details

The queue initialization process:

1. Creates worker channels based on the specified concurrency level
2. Starts goroutines for each worker channel
3. Sets up the job processing pipeline
4. Initializes internal synchronization mechanisms

Note: Users don't need to call `Init()` directly as it's automatically called by the constructor functions (`New()` and `NewPQ()`).

## License

MIT License

## Contributing

Contributions are welcome! Please feel free to submit a Pull Request.
