package queue

import (
	"container/heap"
)

// priorityQueue implements heap.Interface for a slice of *Item[T].
type priorityQueue[T comparable] struct {
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

// removeAt removes the item at index i from the heap, then re-heapifies.
// Returns the removed item.
func (pq *priorityQueue[T]) removeAt(i int) *Item[T] {
	n := pq.Len()

	// Swap the item to remove with the last element.
	pq.Swap(i, n-1)

	// Pop the last element (which is the item we intended to remove).
	popped := heap.Pop(pq).(*Item[T])

	// If i < n-1, we swapped, so we may need to fix the heap property at i
	// The standard approach is to bubble up or down from i.
	if i < n-1 {
		heap.Fix(pq, i) // O(log n)
	}
	return popped
}

// ---------------------------------------------------------------

// PriorityQueue is the user-facing wrapper around priorityQueue[T].
type PriorityQueue[T comparable] struct {
	internal *priorityQueue[T]
}

// NewPriorityQueue initializes an empty priority queue.
func NewPriorityQueue[T comparable]() *PriorityQueue[T] {
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
	q = NewPriorityQueue[T]()
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

// Peek returns the item with the smallest priority without removing it.
func (q *PriorityQueue[T]) Peek() (T, int, bool) {
	var zeroValue T
	if q.internal.Len() == 0 {
		return zeroValue, 0, false
	}
	top := q.internal.items[0]
	return top.Value, top.Priority, true
}

// IsEmpty checks if the queue has no items.
func (q *PriorityQueue[T]) IsEmpty() bool {
	return q.internal.Len() == 0
}
