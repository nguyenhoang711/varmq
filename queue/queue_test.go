package queue

import (
	"fmt"
	"testing"
	"time"
)

func TestQueue_Add(t *testing.T) {
	t.Parallel()
	q := New(2, func(data int) int {
		time.Sleep(100 * time.Millisecond)
		return data * 2
	})
	defer q.Close()

	resultChan, _ := q.Add(5)
	result := <-resultChan

	if result != 10 {
		t.Errorf("Expected result to be 10, got %d", result)
	}
}

func TestQueue_AddAll(t *testing.T) {
	q := New(2, func(data int) int {
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
}

func TestQueue_RemoveJob(t *testing.T) {
	q := New(2, func(data int) int {
		time.Sleep(100 * time.Millisecond)
		return data * 2
	})
	defer q.Close()

	_, cancel := q.Add(5)
	cancel()

	if q.PendingCount() != 0 {
		t.Errorf("Expected pending count to be 0, got %d", q.PendingCount())
	}
}

func TestQueue_WaitUntilFinished(t *testing.T) {
	q := New(2, func(data int) int {
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
}

func TestQueue_Concurrency(t *testing.T) {
	t.Parallel()
	concurrency := uint(2)
	processed := 0
	q := New(concurrency, func(data int) int {
		time.Sleep(100 * time.Millisecond)
		processed++
		return data * 2
	})
	defer q.Close()

	// Add more jobs than concurrency
	for i := 0; i < 5; i++ {
		q.Add(i)
	}

	// Wait a bit to let some processing happen
	time.Sleep(150 * time.Millisecond)

	// Check if only concurrency number of jobs are being processed
	if current := q.CurrentProcessingCount(); current > concurrency {
		t.Errorf("Processing more jobs than concurrency allows. Got %d, want <= %d", current, concurrency)
	}
}

func TestQueue_WaitAndClose(t *testing.T) {
	q := New(2, func(data int) int {
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
}

func BenchmarkQueue_Add(b *testing.B) {
	q := New(20, func(data int) int {
		return data * 2
	})
	defer q.Close()

	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		out, _ := q.Add(i)
		<-out
	}
}

func BenchmarkQueue_AddAll(b *testing.B) {
	q := New(10, func(data int) int {
		return data * 2
	})
	defer q.Close()

	data := make([]int, 1000)
	for i := range data {
		data[i] = i
	}

	b.ResetTimer()
	fmt.Println("Data:", b.N)
	for i := 0; i < b.N; i++ {
		out := q.AddAll(data...)

		for range out {
			// drain the channel
		}
	}
}
