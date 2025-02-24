package utils

import (
	"sync"
	"testing"
)

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

func BenchmarkShortID(b *testing.B) {
	for i := 0; i < b.N; i++ {
		_, _ = ShortID(8)
	}
}
