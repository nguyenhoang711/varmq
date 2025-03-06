package queue

import (
	"testing"

	"github.com/fahimfaisaal/gocq/internal/queue/types"
)

func TestQueue(t *testing.T) {
	t.Run("Basic Operations", func(t *testing.T) {
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
	})

	t.Run("Empty Queue", func(t *testing.T) {
		q := NewQueue[int]()
		val, ok := q.Dequeue()
		if ok || val != 0 {
			t.Errorf("expected no value, got %d", val)
		}
	})

	t.Run("Values Method", func(t *testing.T) {
		q := NewQueue[int]()
		q.Enqueue(types.Item[int]{Value: 1})
		q.Enqueue(types.Item[int]{Value: 2})

		values := q.Values()
		expected := []int{1, 2}
		for i, v := range values {
			if v != expected[i] {
				t.Errorf("expected value %d, got %d", expected[i], v)
			}
		}
	})
}
