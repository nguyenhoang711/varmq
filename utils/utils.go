package utils

import (
	"crypto/rand"
	"encoding/base64"
	"sync"
)

func FanIn[T any](channels []<-chan T) <-chan T {
	wg := new(sync.WaitGroup)
	// We create an output channel to funnel all messages into one place.
	out := make(chan T)

	wg.Add(len(channels))

	// For each channel, start a goroutine that copies items to `out`.
	for _, ch := range channels {
		go func(c <-chan T) {
			defer wg.Done()
			for val := range c {
				out <- val
			}
		}(ch)
	}

	go func() {
		wg.Wait()
		close(out)
	}()

	// Return the merged channel.
	return out
}

func ShortID(length int) (string, error) {
	// Generate random bytes
	b := make([]byte, length)
	if _, err := rand.Read(b); err != nil {
		return "", err
	}

	// Encode to base64 (URL-safe) and trim padding
	return base64.URLEncoding.EncodeToString(b)[:length], nil
}
