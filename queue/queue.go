package queue

import (
	"container/list"
	"sync"
)

type Job[T, R any] struct {
	data     T
	response chan R
}

type Queue[T, R any] struct {
	concurrency   uint
	worker        func(T) R
	channelsStack []chan *Job[T, R]
	curProcessing uint
	jobQueue      *list.List
	wg            *sync.WaitGroup
	mx            *sync.Mutex
}

// Creates a new Queue with the specified concurrency and worker function.
// O(1)
func New[T, R any](concurrency uint, worker func(T) R) *Queue[T, R] {
	channelsStack := make([]chan *Job[T, R], concurrency)
	wg, mx, jobQueue := new(sync.WaitGroup), new(sync.Mutex), new(list.List)

	queue := &Queue[T, R]{concurrency, worker, channelsStack, 0, jobQueue, wg, mx}

	return queue.Init()
}

// Initializes the Queue by starting the worker goroutines.
// O(n) where n is the concurrency
func (q *Queue[T, R]) Init() *Queue[T, R] {
	for i := range q.concurrency {
		q.channelsStack[i] = make(chan *Job[T, R])

		go func(channel chan *Job[T, R]) {
			for job := range channel {
				output := q.worker(job.data)
				job.response <- output // sends the output to the job consumer
				q.wg.Done()
				close(job.response)

				q.mx.Lock()
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
func (q *Queue[T, R]) pickNextChannel() chan<- *Job[T, R] {
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
func (q *Queue[T, R]) PendingCount() int {
	return q.jobQueue.Len()
}

// Returns the number of jobs currently being processed.
// O(1)
func (q *Queue[T, R]) CurrentProcessingCount() uint {
	return q.curProcessing
}

// Adds a new job to the queue and returns a channel to receive the response and a cancel function.
// O(1)
func (q *Queue[T, R]) Add(data T) (<-chan R, func()) {
	job := Job[T, R]{
		data:     data,
		response: make(chan R, 1),
	}

	q.mx.Lock()
	defer q.mx.Unlock()
	el := q.jobQueue.PushBack(&job)
	q.wg.Add(1)

	// process next job only when the current processing job count is less than the concurrency
	if q.curProcessing < q.concurrency {
		q.processNextJob()
	}

	return job.response, func() {
		q.RemoveJob(el)
		close(job.response)
	}
}

// Adds multiple jobs to the queue and returns a channel to receive all responses.
// O(n) where n is the number of jobs added
func (q *Queue[T, R]) AddAll(data ...T) <-chan R {
	wg := new(sync.WaitGroup)
	merged := make(chan R)

	wg.Add(len(data))
	for _, item := range data {
		res, _ := q.Add(item)

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

// Removes a job from the queue by its element.
// O(1)
func (q *Queue[T, R]) RemoveJob(el *list.Element) {
	q.mx.Lock()
	defer q.mx.Unlock()
	q.jobQueue.Remove(el)
	q.wg.Done()
}

// Processes the next job in the queue.
// O(1)
func (q *Queue[T, R]) processNextJob() {
	element := q.jobQueue.Front()

	if element == nil {
		return
	}

	q.jobQueue.Remove(element)
	value, _ := element.Value.(*Job[T, R])

	q.curProcessing++

	go func(data *Job[T, R]) {
		q.pickNextChannel() <- data
	}(value)
}

// Waits until all pending jobs in the queue are processed.
// O(n) where n is the number of pending jobs
func (q *Queue[T, R]) WaitUntilFinished() {
	q.wg.Wait()
}

// Closes the queue and resets all internal states.
// O(n) where n is the number of channels
func (q *Queue[T, R]) Close() {
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
func (q *Queue[T, R]) WaitAndClose() {
	q.WaitUntilFinished()
	q.Close()
}
