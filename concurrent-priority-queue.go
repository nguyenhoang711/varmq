package gocq

import (
	"sync"
	"sync/atomic"

	"github.com/fahimfaisaal/gocq/internal/queue"
	types "github.com/fahimfaisaal/gocq/internal/queue/types"
)

type concurrentPriorityQueue[T, R any] struct {
	*concurrentQueue[T, R]
}

// NewPriorityQueue creates a new concurrentPriorityQueue with the specified concurrency and worker function.
func NewPriorityQueue[T, R any](concurrency uint, worker func(T) R) *concurrentPriorityQueue[T, R] {
	channelsStack := make([]chan *types.Job[T, R], concurrency)
	wg, mx, jobQueue := new(sync.WaitGroup), new(sync.Mutex), queue.NewPriorityQueue[*types.Job[T, R]]()

	queue := &concurrentQueue[T, R]{
		concurrency:   concurrency,
		worker:        worker,
		channelsStack: channelsStack,
		curProcessing: 0,
		jobQueue:      jobQueue,
		wg:            wg,
		mx:            mx,
		isPaused:      atomic.Bool{},
	}

	return &concurrentPriorityQueue[T, R]{concurrentQueue: queue.Init()}
}

// Pause pauses the processing of jobs.
func (q *concurrentPriorityQueue[T, R]) Pause() *concurrentPriorityQueue[T, R] {
	q.isPaused.Store(true)
	return q
}

// Add adds a new Job with the given priority to the queue and returns a channel to receive the response.
// Time complexity: O(log n)
func (q *concurrentPriorityQueue[T, R]) Add(data T, priority int) <-chan R {
	q.mx.Lock()
	defer q.mx.Unlock()

	job := &types.Job[T, R]{
		Data:     data,
		Response: make(chan R, 1),
	}

	q.jobQueue.Enqueue(types.Item[*types.Job[T, R]]{Value: job, Priority: priority})
	q.wg.Add(1)

	// process next Job only when the current processing Job count is less than the concurrency
	if q.shouldProcessNextJob("add") {
		q.processNextJob()
	}

	return job.Response
}

// AddAll adds multiple Jobs with the given priority to the queue and returns a channel to receive all responses.
// Time complexity: O(n log n) where n is the number of Jobs added
func (q *concurrentPriorityQueue[T, R]) AddAll(priority int, data ...T) <-chan R {
	fanIn := withFanIn(func(item T) <-chan R {
		return q.Add(item, priority)
	})
	return fanIn(data...)
}
