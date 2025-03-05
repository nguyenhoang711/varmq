package queue

import "sync"

type gopq[T, R any] struct {
	*gocq[T, R]
}

func NewPQ[T, R any](concurrency uint, worker func(T) R) *gopq[T, R] {
	channelsStack := make([]chan *job[T, R], 0)
	wg, mx, jobQueue := new(sync.WaitGroup), new(sync.Mutex), newPriorityQueue[*job[T, R]]()

	queue := &gocq[T, R]{concurrency, worker, channelsStack, 0, jobQueue, wg, mx}

	return &gopq[T, R]{gocq: queue.Init()}
}

func (q *gopq[T, R]) Add(data T, priority int) <-chan R {
	q.mx.Lock()
	defer q.mx.Unlock()

	j := &job[T, R]{
		data:     data,
		response: make(chan R, 1),
	}

	q.jobQueue.Enqueue(item[*job[T, R]]{Value: j, Priority: priority})
	q.wg.Add(1)

	// process next job only when the current processing job count is less than the concurrency
	if q.curProcessing < q.concurrency {
		q.processNextjob()
	}

	return j.response
}

func (q *gopq[T, R]) AddAll(priority int, data ...T) <-chan R {
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
