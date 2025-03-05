package queue

import (
	"container/heap"
)

// priorityQueue implements heap.Interface for a slice of *Item[T].
type priorityQueue[T any] struct {
	items []*Item[T]
}

// Len, Less, Swap are standard for heap.Interface.

func (pq *priorityQueue[T]) Len() int {
	return len(pq.items)
}

// Less: smaller Priority = higher priority (pops first).
func (pq *priorityQueue[T]) Less(i, j int) bool {
	return pq.items[i].Priority < pq.items[j].Priority
}

// Swap swaps two items in the heap array, updating their indices.
func (pq *priorityQueue[T]) Swap(i, j int) {
	pq.items[i], pq.items[j] = pq.items[j], pq.items[i]
}

// Push is called by heap.Push to add a new element to the end.
func (pq *priorityQueue[T]) Push(x any) {
	pq.items = append(pq.items, x.(*Item[T]))
}

// Pop is called by heap.Pop to remove the last element from the slice.
func (pq *priorityQueue[T]) Pop() any {
	n := len(pq.items)
	item := pq.items[n-1]
	pq.items = pq.items[:n-1]
	return item
}

// PriorityQueue is the user-facing wrapper around priorityQueue[T].
type PriorityQueue[T any] struct {
	internal *priorityQueue[T]
}

// newPriorityQueue initializes an empty priority queue.
func newPriorityQueue[T any]() *PriorityQueue[T] {
	pq := &priorityQueue[T]{
		items: make([]*Item[T], 0),
	}
	heap.Init(pq)
	return &PriorityQueue[T]{internal: pq}
}

// Enqueue pushes a new item with the given priority.
func (q *PriorityQueue[T]) Len() int {
	return q.internal.Len()
}

func (q *PriorityQueue[T]) Init() {
	q.internal.items = make([]*Item[T], 0)
	heap.Init(q.internal)
}

func (q *PriorityQueue[T]) Values() []T {
	values := make([]T, 0)
	for _, item := range q.internal.items {
		values = append(values, item.Value)
	}
	return values
}

// Enqueue pushes a new item with the given priority.
func (q *PriorityQueue[T]) Enqueue(t Item[T]) {
	item := &Item[T]{Value: t.Value, Priority: t.Priority}
	heap.Push(q.internal, item) // O(log n)
}

// Dequeue removes and returns the item with the *smallest* Priority.
func (q *PriorityQueue[T]) Dequeue() (T, bool) {
	var zeroValue T
	if q.internal.Len() == 0 {
		return zeroValue, false
	}
	popped := heap.Pop(q.internal).(*Item[T]) // O(log n)
	return popped.Value, true
}
