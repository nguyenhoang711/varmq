package queue

type job[T, R any] struct {
	data     T
	response chan R
}

// item represents a single element stored in the priority queue.
type item[T any] struct {
	Value    T
	Priority int
}

type iQueue[T any] interface {
	Dequeue() (T, bool)
	Enqueue(item item[T])
	Init()
	Len() int
	Values() []T
}
