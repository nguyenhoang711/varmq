package queue

import (
	"container/list"
)

// queue holds a linked list-based queue of generic items.
type queue[T any] struct {
	internal *list.List
}

// newQueue creates an empty queue using container/list.
func newQueue[T any]() *queue[T] {
	return &queue[T]{internal: new(list.List)}
}

func (q *queue[T]) Init() {
	q.internal.Init()
}

func (q *queue[T]) Values() []T {
	values := make([]T, 0)

	for e := q.internal.Front(); e != nil; e = e.Next() {
		values = append(values, e.Value.(T))
	}

	return values
}

func (q *queue[T]) Len() int {
	return q.internal.Len()
}

// Enqueue adds an item at the back of the list in O(1).
func (q *queue[T]) Enqueue(item item[T]) {
	q.internal.PushBack(item.Value)
}

// Dequeue removes and returns the front item in O(1).
func (q *queue[T]) Dequeue() (T, bool) {
	front := q.internal.Front()
	if front == nil {
		// Return zero value + false if empty
		var zeroValue T
		return zeroValue, false
	}

	// Retrieve the item’s value and remove the element in O(1).
	val := front.Value.(T)
	q.internal.Remove(front)
	return val, true
}
