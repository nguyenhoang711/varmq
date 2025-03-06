package gocq

import (
	"sync"
	"sync/atomic"

	"github.com/fahimfaisaal/gocq/internal/queue"
	types "github.com/fahimfaisaal/gocq/internal/queue/types"
)

type concurrentQueue[T, R any] struct {
	concurrency   uint
	worker        func(T) R
	channelsStack []chan *types.Job[T, R]
	curProcessing uint
	jobQueue      types.IQueue[*types.Job[T, R]]
	wg            *sync.WaitGroup
	mx            *sync.Mutex
	isPaused      atomic.Bool
}

// Creates a new concurrentQueue with the specified concurrency and worker function.
// O(1)
func New[T, R any](concurrency uint, worker func(T) R) *concurrentQueue[T, R] {
	channelsStack := make([]chan *types.Job[T, R], concurrency)
	wg, mx, jobQueue := new(sync.WaitGroup), new(sync.Mutex), queue.NewQueue[*types.Job[T, R]]()

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

	return queue.Init()
}

// Initializes the concurrentQueue by starting the worker goroutines.
// O(n) where n is the concurrency
func (q *concurrentQueue[T, R]) Init() *concurrentQueue[T, R] {
	for i := range q.concurrency {
		// if channel is not nil, close it
		// reason: to avoid routine leaks
		if channel := q.channelsStack[i]; channel != nil {
			close(channel)
		}

		q.channelsStack[i] = make(chan *types.Job[T, R])

		go func(channel chan *types.Job[T, R]) {
			for job := range channel {
				output := q.worker(job.Data)

				job.Response <- output // sends the output to the Job consumer
				close(job.Response)
				q.wg.Done()

				q.mx.Lock()
				// adding the free channel to stack
				q.channelsStack = append(q.channelsStack, channel)
				q.curProcessing--

				// process only if the queue is not empty
				if q.shouldProcessNextJob("next") {
					q.processNextJob()
				}
				q.mx.Unlock()
			}
		}(q.channelsStack[i])
	}

	return q
}

// Picks the next available channel for processing a Job.
// O(1)
func (q *concurrentQueue[T, R]) pickNextChannel() chan<- *types.Job[T, R] {
	q.mx.Lock()
	defer q.mx.Unlock()
	l := len(q.channelsStack)

	// pop the last free channel
	channel := q.channelsStack[l-1]
	q.channelsStack = q.channelsStack[:l-1]
	return channel
}

// Returns the number of Jobs pending in the queue.
// O(1)
func (q *concurrentQueue[T, R]) PendingCount() int {
	return q.jobQueue.Len()
}

func (q *concurrentQueue[T, R]) IsPaused() bool {
	return q.isPaused.Load()
}

// Returns the number of Jobs currently being processed.
// O(1)
func (q *concurrentQueue[T, R]) CurrentProcessingCount() uint {
	return q.curProcessing
}

func (q *concurrentQueue[T, R]) Pause() {
	q.isPaused.Store(true)
}

// Resume continues processing jobs
func (q *concurrentQueue[T, R]) Resume() {
	q.isPaused.Store(false)

	// Process pending jobs if any
	q.mx.Lock()
	defer q.mx.Unlock()

	// Process jobs up to concurrency limit
	for q.shouldProcessNextJob("resume") {
		q.processNextJob()
	}
}

// Adds a new Job to the queue and returns a channel to receive the response and a cancel function.
// O(1)
func (q *concurrentQueue[T, R]) Add(data T) <-chan R {
	q.mx.Lock()
	defer q.mx.Unlock()

	job := &types.Job[T, R]{
		Data:     data,
		Response: make(chan R, 1),
	}

	q.jobQueue.Enqueue(types.Item[*types.Job[T, R]]{Value: job})
	q.wg.Add(1)

	// process next Job only when the current processing Job count is less than the concurrency
	if q.shouldProcessNextJob("add") {
		q.processNextJob()
	}

	return job.Response
}

func (q *concurrentQueue[T, R]) shouldProcessNextJob(from string) bool {
	switch from {
	case "add":
		return !q.isPaused.Load() && q.curProcessing < q.concurrency
	case "resume":
		return q.curProcessing < q.concurrency && q.jobQueue.Len() > 0
	case "next":
		return !q.isPaused.Load() && q.jobQueue.Len() != 0
	default:
		return false
	}
}

// Adds multiple Jobs to the queue and returns a channel to receive all responses.
// O(n) where n is the number of Jobs added
func (q *concurrentQueue[T, R]) AddAll(data ...T) <-chan R {
	fanIn := withFanIn(func(item T) <-chan R {
		return q.Add(item)
	})
	return fanIn(data...)
}

// Processes the next Job in the queue.
// O(1)
func (q *concurrentQueue[T, R]) processNextJob() {
	value, has := q.jobQueue.Dequeue()

	if !has {
		return
	}

	q.curProcessing++

	go func(data *types.Job[T, R]) {
		q.pickNextChannel() <- data
	}(value)
}

// Waits until all pending Jobs in the queue are processed.
// O(n) where n is the number of pending Jobs
func (q *concurrentQueue[T, R]) WaitUntilFinished() {
	q.wg.Wait()
}

// Remove all pending Jobs from the queue.
func (q *concurrentQueue[T, R]) Purge() {
	q.mx.Lock()
	defer q.mx.Unlock()

	prevValues := q.jobQueue.Values()
	q.jobQueue.Init()
	q.wg.Add(-len(prevValues))

	// close all pending channels to avoid routine leaks
	for _, job := range prevValues {
		close(job.Response)
	}
}

// Closes the queue and resets all internal states.
// O(n) where n is the number of channels
func (q *concurrentQueue[T, R]) Close() {
	q.Purge()

	// wait until all ongoing processes are done
	q.wg.Wait()

	for _, channel := range q.channelsStack {
		if channel == nil {
			continue
		}

		close(channel)
	}

	q.channelsStack = make([]chan *types.Job[T, R], q.concurrency)
}

// Waits until all pending Jobs in the queue are processed and then closes the queue.
// O(n) where n is the number of pending Jobs
func (q *concurrentQueue[T, R]) WaitAndClose() {
	q.wg.Wait()
	q.Close()
}
