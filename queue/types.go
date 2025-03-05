package queue

type Job[T, R any] struct {
	data     T
	response chan R
}

// Item represents a single element stored in the priority queue.
type Item[T any] struct {
	Value    T
	Priority int
}

type IQueue[T any] interface {
	Dequeue() (T, bool)
	Enqueue(item Item[T])
	Init()
	Len() int
	Values() []T
}
