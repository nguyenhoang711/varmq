package queue

type Job[T, R any] struct {
	data     T
	response chan R
}

// Item represents a single element stored in the priority queue.
type Item[T comparable] struct {
	Value    T
	Priority int
}

type IQueue[T comparable] interface {
	Dequeue() (T, bool)
	Enqueue(t Item[T])
	Init()
	Len() int
}
