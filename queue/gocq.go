package queue

import (
	"fmt"
	"sync"
)

type Gocq[T, R comparable] struct {
	concurrency   uint
	worker        func(T) R
	channelsStack []chan *Job[T, R]
	curProcessing uint
	jobQueue      IQueue[*Job[T, R]]
	wg            *sync.WaitGroup
	mx            *sync.Mutex
}

type Gopq[T, R comparable] struct {
	*Gocq[T, R]
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

func NewPQ[T, R comparable](concurrency uint, worker func(T) R) *Gopq[T, R] {
	channelsStack := make([]chan *Job[T, R], concurrency)
	wg, mx, jobQueue := new(sync.WaitGroup), new(sync.Mutex), NewPriorityQueue[*Job[T, R]]()

	queue := &Gocq[T, R]{concurrency, worker, channelsStack, 0, jobQueue, wg, mx}

	return &Gopq[T, R]{Gocq: queue}
}

// Creates a new Gocq with the specified concurrency and worker function.
// O(1)
func New[T, R comparable](concurrency uint, worker func(T) R) *Gocq[T, R] {
	channelsStack := make([]chan *Job[T, R], concurrency)
	wg, mx, jobQueue := new(sync.WaitGroup), new(sync.Mutex), NewQueue[*Job[T, R]]()

	fmt.Println(jobQueue.Len())
	queue := &Gocq[T, R]{concurrency, worker, channelsStack, 0, jobQueue, wg, mx}

	return queue.Init()
}

// Initializes the Gocq by starting the worker goroutines.
// O(n) where n is the concurrency
func (q *Gocq[T, R]) Init() *Gocq[T, R] {
	for i := range q.concurrency {
		q.channelsStack[i] = make(chan *Job[T, R])

		go func(channel chan *Job[T, R]) {
			for job := range channel {
				q.mx.Lock()
				q.mx.Unlock()

				output := q.worker(job.data)

				q.mx.Lock()
				job.response <- output // sends the output to the job consumer
				close(job.response)
				q.wg.Done()

				// adding the free channel to stack
				q.channelsStack = append(q.channelsStack, channel)
				q.curProcessing--

				// process only if the queue is not empty
				if q.jobQueue.Len() != 0 {
					q.processNextJob()
				}
				q.mx.Unlock()
			}
		}(q.channelsStack[i])
	}

	return q
}

// Picks the next available channel for processing a job.
// O(1)
func (q *Gocq[T, R]) pickNextChannel() chan<- *Job[T, R] {
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
func (q *Gocq[T, R]) PendingCount() int {
	return q.jobQueue.Len()
}

// Returns the number of jobs currently being processed.
// O(1)
func (q *Gocq[T, R]) CurrentProcessingCount() uint {
	return q.curProcessing
}

// Adds a new job to the queue and returns a channel to receive the response and a cancel function.
// O(1)
func (q *Gocq[T, R]) Add(data T) <-chan R {
	q.mx.Lock()
	defer q.mx.Unlock()

	job := &Job[T, R]{
		data:     data,
		response: make(chan R, 1),
	}

	q.jobQueue.Enqueue(Item[*Job[T, R]]{Value: job})
	q.wg.Add(1)

	// process next job only when the current processing job count is less than the concurrency
	if q.curProcessing < q.concurrency {
		q.processNextJob()
	}

	return job.response
}

// Adds multiple jobs to the queue and returns a channel to receive all responses.
// O(n) where n is the number of jobs added
func (q *Gocq[T, R]) AddAll(data ...T) <-chan R {
	wg := new(sync.WaitGroup)
	merged := make(chan R)

	wg.Add(len(data))
	for _, item := range data {
		res := q.Add(item)

		go func(c <-chan R) {
			defer wg.Done()
			for val := range c {
				merged <- val
			}
		}(res)
	}

	go func() {
		wg.Wait()
		close(merged)
	}()

	return merged
}

// Processes the next job in the queue.
// O(1)
func (q *Gocq[T, R]) processNextJob() {
	value, has := q.jobQueue.Dequeue()

	if !has {
		return
	}

	q.curProcessing++

	go func(data *Job[T, R]) {
		q.pickNextChannel() <- data
	}(value)
}

// Waits until all pending jobs in the queue are processed.
// O(n) where n is the number of pending jobs
func (q *Gocq[T, R]) WaitUntilFinished() {
	q.wg.Wait()
}

// Closes the queue and resets all internal states.
// O(n) where n is the number of channels
func (q *Gocq[T, R]) Close() {
	// reset all stores
	q.mx.Lock()
	pendingCount := q.PendingCount()
	q.jobQueue.Init()
	q.wg.Add(-pendingCount)
	q.mx.Unlock()

	// wait until all ongoing processes are done
	q.wg.Wait()

	for _, channel := range q.channelsStack {
		close(channel)
	}

	q.channelsStack = make([]chan *Job[T, R], q.concurrency)
}

// Waits until all pending jobs in the queue are processed and then closes the queue.
// O(n) where n is the number of pending jobs
func (q *Gocq[T, R]) WaitAndClose() {
	q.WaitUntilFinished()
	q.Close()
}
