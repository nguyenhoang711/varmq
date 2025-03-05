package queue

import (
	"testing"

	"github.com/fahimfaisaal/gocq/internal/queue/types"
)

func TestQueue(t *testing.T) {
	q := NewQueue[int]()

	if q.Len() != 0 {
		t.Errorf("expected length 0, got %d", q.Len())
	}

	q.Enqueue(types.Item[int]{Value: 1})
	q.Enqueue(types.Item[int]{Value: 2})

	if q.Len() != 2 {
		t.Errorf("expected length 2, got %d", q.Len())
	}

	val, ok := q.Dequeue()
	if !ok || val != 1 {
		t.Errorf("expected value 1, got %d", val)
	}

	val, ok = q.Dequeue()
	if !ok || val != 2 {
		t.Errorf("expected value 2, got %d", val)
	}

	val, ok = q.Dequeue()
	if ok || val != 0 {
		t.Errorf("expected no value, got %d", val)
	}
}
