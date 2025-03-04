package queue

import (
	"container/list"
)

// Queue holds a linked list-based queue of generic items.
type Queue[T comparable] struct {
	internal *list.List
}

// newQueue creates an empty Queue using container/list.
func NewQueue[T comparable]() *Queue[T] {
	return &Queue[T]{internal: new(list.List)}
}

func (q *Queue[T]) Init() {
	q.internal.Init()
}

func (q *Queue[T]) Len() int {
	return q.internal.Len()
}

// Enqueue adds an item at the back of the list in O(1).
func (q *Queue[T]) Enqueue(item Item[T]) {
	q.internal.PushBack(item.Value)
}

// Dequeue removes and returns the front item in O(1).
func (q *Queue[T]) Dequeue() (T, bool) {
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

// Remove finds the first occurrence of 'value' in the queue and removes it.
// Returns true if an item was removed, false if the item wasn't in the queue.
func (q *Queue[T]) Remove(value T) bool {
	for e := q.internal.Front(); e != nil; e = e.Next() {
		if e.Value.(T) == value {
			q.internal.Remove(e)
			return true
		}
	}
	return false
}
