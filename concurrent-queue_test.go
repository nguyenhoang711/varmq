package gocq

import (
	"testing"
	"time"
)

func TestConcurrentQueue(t *testing.T) {
	t.Run("Add", func(t *testing.T) {
		t.Parallel()
		q := NewQueue(2, func(data int) int {
			time.Sleep(100 * time.Millisecond)
			return data * 2
		})
		defer q.Close()

		resultChan := q.Add(5)
		result := <-resultChan

		if result != 10 {
			t.Errorf("Expected result to be 10, got %d", result)
		}
	})

	t.Run("AddAll", func(t *testing.T) {
		q := NewQueue(2, func(data int) int {
			time.Sleep(100 * time.Millisecond)
			return data * 2
		})
		defer q.Close()

		data := []int{1, 2, 3, 4, 5}
		resultChan := q.AddAll(data...)

		results := make([]int, 0)
		for result := range resultChan {
			results = append(results, result)
		}

		if len(results) != 5 {
			t.Errorf("Expected 5 results, got %d", len(results))
		}

		expectedSum := 30 // sum of [2,4,6,8,10]
		actualSum := 0
		for _, r := range results {
			actualSum += r
		}

		if actualSum != expectedSum {
			t.Errorf("Expected sum of results to be %d, got %d", expectedSum, actualSum)
		}
	})

	t.Run("WaitUntilFinished", func(t *testing.T) {
		q := NewQueue(2, func(data int) int {
			time.Sleep(100 * time.Millisecond)
			return data * 2
		})
		defer q.Close()

		q.Add(1)
		q.Add(2)
		q.Add(3)

		q.WaitUntilFinished()

		if q.PendingCount() != 0 {
			t.Errorf("Expected pending count to be 0, got %d", q.PendingCount())
		}
		if q.CurrentProcessingCount() != 0 {
			t.Errorf("Expected current processing count to be 0, got %d", q.CurrentProcessingCount())
		}
	})

	t.Run("Concurrency", func(t *testing.T) {
		t.Parallel()
		concurrency := uint(2)
		processed := 0
		q := NewQueue(concurrency, func(data int) int {
			time.Sleep(100 * time.Millisecond)
			processed++
			return data * 2
		})
		defer q.Close()

		// Add more Jobs than concurrency
		for i := 0; i < 5; i++ {
			q.Add(i)
		}

		// Wait a bit to let some processing happen
		time.Sleep(150 * time.Millisecond)

		// Check if only concurrency number of Jobs are being processed
		if current := q.CurrentProcessingCount(); current > concurrency {
			t.Errorf("Processing more Jobs than concurrency allows. Got %d, want <= %d", current, concurrency)
		}
	})

	t.Run("WaitAndClose", func(t *testing.T) {
		q := NewQueue(2, func(data int) int {
			time.Sleep(100 * time.Millisecond)
			return data * 2
		})

		q.Add(1)
		q.Add(2)
		q.Add(3)

		q.WaitAndClose()

		if q.PendingCount() != 0 {
			t.Errorf("Expected pending count to be 0, got %d", q.PendingCount())
		}
		if q.CurrentProcessingCount() != 0 {
			t.Errorf("Expected current processing count to be 0, got %d", q.CurrentProcessingCount())
		}
	})

	t.Run("PauseAndResume", func(t *testing.T) {
		worker := func(data int) int {
			time.Sleep(100 * time.Millisecond)
			return data * 2
		}

		q := NewQueue(2, worker)

		resp1 := q.Add(1)
		resp2 := q.Add(2)

		if result := <-resp1; result != 2 {
			t.Errorf("Expected result to be 2, got %d", result)
		}
		if result := <-resp2; result != 4 {
			t.Errorf("Expected result to be 4, got %d", result)
		}

		q.Pause()
		resp3 := q.Add(3)
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
	})
}

func BenchmarkQueue_Add(b *testing.B) {
	q := NewQueue(20, func(data int) int {
		return data * 2
	})
	defer q.Close()

	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		out := q.Add(i)
		<-out
	}
}

func BenchmarkQueue_AddAll(b *testing.B) {
	q := NewQueue(10, func(data int) int {
		return data * 2
	})
	defer q.Close()

	data := make([]int, 1000)
	for i := range data {
		data[i] = i
	}

	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		out := q.AddAll(data...)

		for range out {
			// drain the channel
		}
	}
}

func BenchmarkQueue_WaitUntilFinished(b *testing.B) {
	q := NewQueue(20, func(data int) int {
		return data * 2
	})
	defer q.Close()

	for i := 0; i < 1000; i++ {
		q.Add(i)
	}

	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		q.WaitUntilFinished()
	}
}

func BenchmarkQueue_WaitAndClose(b *testing.B) {
	q := NewQueue(20, func(data int) int {
		return data * 2
	})

	for i := 0; i < 1000; i++ {
		q.Add(i)
	}

	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		q.WaitAndClose()
	}
}
