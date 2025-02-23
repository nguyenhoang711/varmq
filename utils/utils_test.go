package utils

import (
	"sync"
	"testing"
)

func TestFanIn(t *testing.T) {
	t.Run("should merge multiple channels into one", func(t *testing.T) {
		// Create test channels
		ch1 := make(chan int, 2)
		ch2 := make(chan int, 2)
		ch3 := make(chan int, 2)

		// Send test data
		ch1 <- 1
		ch1 <- 2
		ch2 <- 3
		ch2 <- 4
		ch3 <- 5
		ch3 <- 6

		// Close all channels
		close(ch1)
		close(ch2)
		close(ch3)

		// Fan in the channels
		merged := FanIn([]<-chan int{ch1, ch2, ch3})

		// Collect results
		results := make([]int, 0, 6)
		for n := range merged {
			results = append(results, n)
		}

		// Verify length
		if len(results) != 6 {
			t.Errorf("Expected 6 results, got %d", len(results))
		}

		// Verify sum (order doesn't matter)
		expectedSum := 21 // 1+2+3+4+5+6
		actualSum := 0
		for _, v := range results {
			actualSum += v
		}

		if actualSum != expectedSum {
			t.Errorf("Expected sum %d, got %d", expectedSum, actualSum)
		}
	})

	t.Run("should handle empty channels list", func(t *testing.T) {
		merged := FanIn([]<-chan int{})
		count := 0
		for range merged {
			count++
		}
		if count != 0 {
			t.Errorf("Expected 0 items from empty channels, got %d", count)
		}
	})
}

func TestShortID(t *testing.T) {
	t.Run("should generate ID of correct length", func(t *testing.T) {
		length := 5
		id, err := ShortID(length)
		if err != nil {
			t.Errorf("Unexpected error: %v", err)
		}
		if len(id) != length {
			t.Errorf("Expected length %d, got %d", length, len(id))
		}
	})

	t.Run("should generate unique IDs", func(t *testing.T) {
		iterations := 1000
		length := 8
		ids := make(map[string]bool)
		var mx sync.Mutex
		var wg sync.WaitGroup

		for i := 0; i < iterations; i++ {
			wg.Add(1)
			go func() {
				defer wg.Done()
				id, err := ShortID(length)
				if err != nil {
					t.Errorf("Unexpected error: %v", err)
					return
				}

				mx.Lock()
				if ids[id] {
					t.Errorf("Duplicate ID generated: %s", id)
				}
				ids[id] = true
				mx.Unlock()
			}()
		}

		wg.Wait()

		if len(ids) != iterations {
			t.Errorf("Expected %d unique IDs, got %d", iterations, len(ids))
		}
	})

	t.Run("should handle invalid length", func(t *testing.T) {
		_, err := ShortID(0)
		if err != nil {
			t.Errorf("Expected no error for length 0, got %v", err)
		}
	})
}

func BenchmarkFanIn(b *testing.B) {
	numChannels := 100
	channels := make([]<-chan int, numChannels)

	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		b.StopTimer()
		// Setup channels
		for j := range channels {
			ch := make(chan int, 1)
			ch <- j
			close(ch)
			channels[j] = ch
		}

		b.StartTimer()
		merged := FanIn(channels)
		for range merged {
			// drain the channel
		}
	}
}

func BenchmarkShortID(b *testing.B) {
	for i := 0; i < b.N; i++ {
		_, _ = ShortID(8)
	}
}
