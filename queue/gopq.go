package queue

import "sync"

type Gopq[T, R comparable] struct {
	*Gocq[T, R]
}

func NewPQ[T, R comparable](concurrency uint, worker func(T) R) *Gopq[T, R] {
	channelsStack := make([]chan *Job[T, R], 0)
	wg, mx, jobQueue := new(sync.WaitGroup), new(sync.Mutex), newPriorityQueue[*Job[T, R]]()

	queue := &Gocq[T, R]{concurrency, worker, channelsStack, 0, jobQueue, wg, mx}

	return &Gopq[T, R]{Gocq: queue.Init()}
}

func (q *Gopq[T, R]) Add(data T, priority int) <-chan R {
	q.mx.Lock()
	defer q.mx.Unlock()

	job := &Job[T, R]{
		data:     data,
		response: make(chan R, 1),
	}

	q.jobQueue.Enqueue(Item[*Job[T, R]]{Value: job, Priority: priority})
	q.wg.Add(1)

	// process next job only when the current processing job count is less than the concurrency
	if q.curProcessing < q.concurrency {
		q.processNextJob()
	}

	return job.response
}

func (q *Gopq[T, R]) AddAll(priority int, data ...T) <-chan R {
	fanIn := WithFanIn(func(item T) <-chan R {
		return q.Add(item, priority)
	})
	return fanIn(data...)
}

func WithFanIn[T, R any](fn func(T) <-chan R) func(...T) <-chan R {
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
