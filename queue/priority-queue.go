package queue

import (
	"container/heap"
)

// pQueue implements heap.Interface for a slice of *item[T].
type pQueue[T any] struct {
	items []*item[T]
}

// Len, Less, Swap are standard for heap.Interface.

func (pq *pQueue[T]) Len() int {
	return len(pq.items)
}

// Less: smaller Priority = higher priority (pops first).
func (pq *pQueue[T]) Less(i, j int) bool {
	return pq.items[i].Priority < pq.items[j].Priority
}

// Swap swaps two items in the heap array.
func (pq *pQueue[T]) Swap(i, j int) {
	pq.items[i], pq.items[j] = pq.items[j], pq.items[i]
}

// Push is called by heap.Push to add a new element to the end.
func (pq *pQueue[T]) Push(x any) {
	pq.items = append(pq.items, x.(*item[T]))
}

// Pop is called by heap.Pop to remove the last element from the slice.
func (pq *pQueue[T]) Pop() any {
	n := len(pq.items)
	item := pq.items[n-1]
	pq.items = pq.items[:n-1]
	return item
}

// priorityQueue is the user-facing wrapper around pQueue[T].
type priorityQueue[T any] struct {
	internal *pQueue[T]
}

// newPriorityQueue initializes an empty priority queue.
func newPriorityQueue[T any]() *priorityQueue[T] {
	pq := &pQueue[T]{
		items: make([]*item[T], 0),
	}
	heap.Init(pq)
	return &priorityQueue[T]{internal: pq}
}

// Enqueue pushes a new item with the given priority.
func (q *priorityQueue[T]) Len() int {
	return q.internal.Len()
}

func (q *priorityQueue[T]) Init() {
	q.internal.items = make([]*item[T], 0)
	heap.Init(q.internal)
}

func (q *priorityQueue[T]) Values() []T {
	values := make([]T, 0)
	for _, item := range q.internal.items {
		values = append(values, item.Value)
	}
	return values
}

// Enqueue pushes a new item with the given priority.
func (q *priorityQueue[T]) Enqueue(t item[T]) {
	item := &item[T]{Value: t.Value, Priority: t.Priority}
	heap.Push(q.internal, item) // O(log n)
}

// Dequeue removes and returns the item with the *smallest* Priority.
func (q *priorityQueue[T]) Dequeue() (T, bool) {
	var zeroValue T
	if q.internal.Len() == 0 {
		return zeroValue, false
	}
	popped := heap.Pop(q.internal).(*item[T]) // O(log n)
	return popped.Value, true
}
