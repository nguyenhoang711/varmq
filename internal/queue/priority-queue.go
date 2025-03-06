package queue

import (
	"container/heap"

	types "github.com/fahimfaisaal/gocq/internal/queue/types"
)

// PriorityQueue is the user-facing wrapper around heapQueue[T].
type PriorityQueue[T any] struct {
	internal       *heapQueue[T]
	insertionCount int
}

// newPriorityQueue initializes an empty priority queue.
func NewPriorityQueue[T any]() *PriorityQueue[T] {
	pq := &heapQueue[T]{
		items: make([]*types.Item[T], 0),
	}
	heap.Init(pq)
	return &PriorityQueue[T]{internal: pq}
}

// Enqueue pushes a new item with the given priority.
func (q *PriorityQueue[T]) Len() int {
	return q.internal.Len()
}

func (q *PriorityQueue[T]) Init() {
	q.internal.items = make([]*types.Item[T], 0)
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
func (q *PriorityQueue[T]) Enqueue(t types.Item[T]) {
	t.Index = q.insertionCount
	q.insertionCount++
	heap.Push(q.internal, &t) // O(log n)
}

// Dequeue removes and returns the item with the *smallest* Priority.
func (q *PriorityQueue[T]) Dequeue() (T, bool) {
	var zeroValue T
	if q.internal.Len() == 0 {
		return zeroValue, false
	}
	popped := heap.Pop(q.internal).(*types.Item[T]) // O(log n)
	return popped.Value, true
}
