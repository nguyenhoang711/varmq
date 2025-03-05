package gocq

import (
	"testing"
	"time"
)

func TestConcurrentPriorityQueue(t *testing.T) {
	worker := func(data int) int {
		time.Sleep(100 * time.Millisecond)
		return data * 2
	}

	q := NewPriorityQueue(2, worker)

	resp1 := q.Add(1, 2)
	resp2 := q.Add(2, 1)

	if result := <-resp2; result != 4 {
		t.Errorf("Expected result to be 4, got %d", result)
	}
	if result := <-resp1; result != 2 {
		t.Errorf("Expected result to be 2, got %d", result)
	}

	q.Pause()
	resp3 := q.Add(3, 1)
	time.Sleep(50 * time.Millisecond)
	if count := q.PendingCount(); count != 1 {
		t.Errorf("Expected pending count to be 1, got %d", count)
	}

	q.Resume()
	if result := <-resp3; result != 6 {
		t.Errorf("Expected result to be 6, got %d", result)
	}

	q.WaitUntilFinished()
	if count := q.PendingCount(); count != 0 {
		t.Errorf("Expected pending count to be 0, got %d", count)
	}
	if count := q.CurrentProcessingCount(); count != 0 {
		t.Errorf("Expected current processing count to be 0, got %d", count)
	}
}
