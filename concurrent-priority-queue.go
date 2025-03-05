package gocq

import (
	"sync"

	"github.com/fahimfaisaal/gocq/internal/queue"
	types "github.com/fahimfaisaal/gocq/internal/queue/types"
)

type concurrentPriorityQueue[T, R any] struct {
	*concurrentQueue[T, R]
}

func NewPQ[T, R any](concurrency uint, worker func(T) R) *concurrentPriorityQueue[T, R] {
	channelsStack := make([]chan *types.Job[T, R], 0)
	wg, mx, jobQueue := new(sync.WaitGroup), new(sync.Mutex), queue.NewPriorityQueue[*types.Job[T, R]]()

	queue := &concurrentQueue[T, R]{concurrency, worker, channelsStack, 0, jobQueue, wg, mx}

	return &concurrentPriorityQueue[T, R]{concurrentQueue: queue.Init()}
}

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
	if q.curProcessing < q.concurrency {
		q.processNextJob()
	}

	return job.Response
}

func (q *concurrentPriorityQueue[T, R]) AddAll(priority int, data ...T) <-chan R {
	fanIn := withFanIn(func(item T) <-chan R {
		return q.Add(item, priority)
	})
	return fanIn(data...)
}

func withFanIn[T, R any](fn func(T) <-chan R) func(...T) <-chan R {
	return func(data ...T) <-chan R {
		wg := new(sync.WaitGroup)
		merged := make(chan R)

		wg.Add(len(data))
		for _, item := range data {
			go func(c <-chan R) {
				defer wg.Done()
				for val := range c {
					merged <- val
				}
			}(fn(item))
		}

		go func() {
			wg.Wait()
			close(merged)
		}()

		return merged
	}
}
