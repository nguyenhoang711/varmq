package queue

import (
	"container/list"
	"sync"

	"github.com/fahimfaisaal/gocq/utils"
)

type JobId string

type Job[T, R any] struct {
	id       JobId
	data     T
	response chan R
}

type Queue[T, R any] struct {
	concurrency   uint
	worker        func(T) R
	channels      []chan *Job[T, R]
	jobsMap       map[JobId]*list.Element
	curProcessing uint
	jobQueue      *list.List
	wg            *sync.WaitGroup
	mx            *sync.Mutex
}

func New[T, R any](concurrency uint, worker func(T) R) *Queue[T, R] {
	channels, jobsMap := make([]chan *Job[T, R], concurrency), make(map[JobId]*list.Element)
	wg, mx, jobQueue := new(sync.WaitGroup), new(sync.Mutex), new(list.List)

	queue := &Queue[T, R]{concurrency, worker, channels, jobsMap, 0, jobQueue, wg, mx}

	return queue.Init()
}

func (q *Queue[T, R]) Init() *Queue[T, R] {
	for i := range q.concurrency {
		q.channels[i] = make(chan *Job[T, R])

		go func(channel chan *Job[T, R]) {
			for job := range channel {
				output := q.worker(job.data)
				job.response <- output // sends the output to the job consumer
				q.wg.Done()
				close(job.response)

				q.mx.Lock()
				// adding the free channel to stack
				q.channels = append(q.channels, channel)

				// process only if the queue is not empty
				if q.jobQueue.Len() != 0 {
					q.processNextJob()
				}

				q.curProcessing--
				q.mx.Unlock()

			}
		}(q.channels[i])
	}

	return q
}

// Picking the next free channel from the stack
func (q *Queue[T, R]) pickNextChannel() chan<- *Job[T, R] {
	q.mx.Lock()
	l := len(q.channels)

	// pop the last free channel
	channel := q.channels[l-1]
	q.channels = q.channels[:l-1]
	q.mx.Unlock()
	return channel
}

func (q *Queue[T, R]) PendingCount() int {
	return q.jobQueue.Len()
}

func (q *Queue[T, R]) CurrentProcessingCount() uint {
	return q.curProcessing
}

func (q *Queue[T, R]) Add(data T) (JobId, <-chan R) {
	jobId, _ := utils.ShortID(5)

	job := Job[T, R]{
		id:       JobId(jobId),
		data:     data,
		response: make(chan R, 1),
	}

	q.wg.Add(1)

	q.mx.Lock()
	el := q.jobQueue.PushBack(&job)
	q.jobsMap[job.id] = el

	// process next job only when the current processing job count is less than the concurrency
	if uint(q.jobQueue.Len())+q.curProcessing <= q.concurrency {
		q.processNextJob()
	}
	q.mx.Unlock()

	return job.id, job.response
}

func (q *Queue[T, R]) AddAll(data ...T) <-chan R {
	outs := make([]<-chan R, 0)

	for _, item := range data {
		_, out := q.Add(item)

		q.mx.Lock()
		outs = append(outs, out)
		q.mx.Unlock()
	}

	return utils.FanIn(outs)
}

func (q *Queue[T, R]) RemoveJob(id JobId) {
	q.mx.Lock()
	el, has := q.jobsMap[id]
	if !has {
		q.mx.Unlock()
		return
	}

	job, _ := el.Value.(*Job[T, R])
	q.jobQueue.Remove(el)
	close(job.response)
	q.jobsMap[id] = nil
	delete(q.jobsMap, id)
	q.wg.Done()
	q.mx.Unlock()
}

func (q *Queue[T, R]) processNextJob() {
	element := q.jobQueue.Front()

	if element == nil {
		q.mx.Unlock()
		return
	}

	q.jobQueue.Remove(element)
	value, _ := element.Value.(*Job[T, R])

	// removing job from the map since its removed from the queue
	q.jobsMap[value.id] = nil
	delete(q.jobsMap, value.id)
	q.curProcessing++

	go func(data *Job[T, R]) {
		q.pickNextChannel() <- data
	}(value)
}

func (q *Queue[T, R]) WaitUntilFinished() {
	q.wg.Wait()
}

func (q *Queue[T, R]) Close() {
	// reset all stores
	q.mx.Lock()
	pendingCount := q.PendingCount()
	q.jobQueue.Init()
	q.jobsMap = make(map[JobId]*list.Element)
	q.wg.Add(-pendingCount)
	q.mx.Unlock()

	// wait until all on going processes are done
	q.wg.Wait()

	for _, channel := range q.channels {
		close(channel)
	}
}
