package queue

import (
	"testing"

	"github.com/fahimfaisaal/gocq/internal/queue/types"
)

func TestPriorityQueue(t *testing.T) {
	t.Run("Basic Operations", func(t *testing.T) {
		pq := NewPriorityQueue[int]()

		if pq.Len() != 0 {
			t.Errorf("expected length 0, got %d", pq.Len())
		}

		pq.Enqueue(types.Item[int]{Value: 1, Priority: 2})
		pq.Enqueue(types.Item[int]{Value: 2, Priority: 1})

		if pq.Len() != 2 {
			t.Errorf("expected length 2, got %d", pq.Len())
		}

		val, ok := pq.Dequeue()
		if !ok || val != 2 {
			t.Errorf("expected value 2, got %d", val)
		}

		val, ok = pq.Dequeue()
		if !ok || val != 1 {
			t.Errorf("expected value 1, got %d", val)
		}

		val, ok = pq.Dequeue()
		if ok || val != 0 {
			t.Errorf("expected no value, got %d", val)
		}
	})

	t.Run("Empty Queue", func(t *testing.T) {
		pq := NewPriorityQueue[int]()
		val, ok := pq.Dequeue()
		if ok || val != 0 {
			t.Errorf("expected no value, got %d", val)
		}
	})

	t.Run("Values Method", func(t *testing.T) {
		pq := NewPriorityQueue[int]()
		pq.Enqueue(types.Item[int]{Value: 1, Priority: 2})
		pq.Enqueue(types.Item[int]{Value: 2, Priority: 1})

		values := pq.Values()
		expected := []int{2, 1}
		for i, v := range values {
			if v != expected[i] {
				t.Errorf("expected value %d, got %d", expected[i], v)
			}
		}
	})
}
