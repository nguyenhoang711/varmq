package types

type IQueue[T any] interface {
	Dequeue() (T, bool)
	Enqueue(item Item[T])
	Init()
	Len() int
	Values() []T
}
