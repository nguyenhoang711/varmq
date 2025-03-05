package queue

import (
	"sync"
)

type gocq[T, R any] struct {
	concurrency   uint
	worker        func(T) R
	channelsStack []chan *job[T, R]
	curProcessing uint
	jobQueue      iQueue[*job[T, R]]
	wg            *sync.WaitGroup
	mx            *sync.Mutex
}

// Creates a new gocq with the specified concurrency and worker function.
// O(1)
func New[T, R any](concurrency uint, worker func(T) R) *gocq[T, R] {
	channelsStack := make([]chan *job[T, R], 0)
	wg, mx, jobQueue := new(sync.WaitGroup), new(sync.Mutex), newQueue[*job[T, R]]()

	queue := &gocq[T, R]{concurrency, worker, channelsStack, 0, jobQueue, wg, mx}

	return queue.Init()
}

// Initializes the gocq by starting the worker goroutines.
// O(n) where n is the concurrency
func (q *gocq[T, R]) Init() *gocq[T, R] {
	for i := range q.concurrency {
		q.channelsStack = append(q.channelsStack, make(chan *job[T, R]))

		go func(channel chan *job[T, R]) {
			for job := range channel {
				output := q.worker(job.data)

				job.response <- output // sends the output to the job consumer
				close(job.response)
				q.wg.Done()

				q.mx.Lock()
				// adding the free channel to stack
				q.channelsStack = append(q.channelsStack, channel)
				q.curProcessing--

				// process only if the queue is not empty
				if q.jobQueue.Len() != 0 {
					q.processNextjob()
				}
				q.mx.Unlock()
			}
		}(q.channelsStack[i])
	}

	return q
}

// Picks the next available channel for processing a job.
// O(1)
func (q *gocq[T, R]) pickNextChannel() chan<- *job[T, R] {
	q.mx.Lock()
	defer q.mx.Unlock()
	l := len(q.channelsStack)

	// pop the last free channel
	channel := q.channelsStack[l-1]
	q.channelsStack = q.channelsStack[:l-1]
	return channel
}

// Returns the number of jobs pending in the queue.
// O(1)
func (q *gocq[T, R]) PendingCount() int {
	return q.jobQueue.Len()
}

// Returns the number of jobs currently being processed.
// O(1)
func (q *gocq[T, R]) CurrentProcessingCount() uint {
	return q.curProcessing
}

// Adds a new job to the queue and returns a channel to receive the response and a cancel function.
// O(1)
func (q *gocq[T, R]) Add(data T) <-chan R {
	q.mx.Lock()
	defer q.mx.Unlock()

	j := &job[T, R]{
		data:     data,
		response: make(chan R, 1),
	}

	q.jobQueue.Enqueue(item[*job[T, R]]{Value: j})
	q.wg.Add(1)

	// process next job only when the current processing job count is less than the concurrency
	if q.curProcessing < q.concurrency {
		q.processNextjob()
	}

	return j.response
}

// Adds multiple jobs to the queue and returns a channel to receive all responses.
// O(n) where n is the number of jobs added
func (q *gocq[T, R]) AddAll(data ...T) <-chan R {
	fanIn := withFanIn(func(item T) <-chan R {
		return q.Add(item)
	})
	return fanIn(data...)
}

// Processes the next job in the queue.
// O(1)
func (q *gocq[T, R]) processNextjob() {
	value, has := q.jobQueue.Dequeue()

	if !has {
		return
	}

	q.curProcessing++

	go func(data *job[T, R]) {
		q.pickNextChannel() <- data
	}(value)
}

// Waits until all pending jobs in the queue are processed.
// O(n) where n is the number of pending jobs
func (q *gocq[T, R]) WaitUntilFinished() {
	q.wg.Wait()
}

// Remove all pending jobs from the queue.
func (q *gocq[T, R]) Purge() {
	q.mx.Lock()
	defer q.mx.Unlock()

	prevValues := q.jobQueue.Values()
	q.jobQueue.Init()
	q.wg.Add(-len(prevValues))

	// close all pending channels to avoid routine leaks
	for _, job := range prevValues {
		close(job.response)
	}
}

// Closes the queue and resets all internal states.
// O(n) where n is the number of channels
func (q *gocq[T, R]) Close() {
	q.Purge()

	// wait until all ongoing processes are done
	q.wg.Wait()

	for _, channel := range q.channelsStack {
		close(channel)
	}

	q.channelsStack = make([]chan *job[T, R], 0)
}

// Waits until all pending jobs in the queue are processed and then closes the queue.
// O(n) where n is the number of pending jobs
func (q *gocq[T, R]) WaitAndClose() {
	q.wg.Wait()
	q.Close()
}
