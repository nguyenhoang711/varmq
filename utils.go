package gocq

import "sync"

func withFanIn[T, R any](fn func(T) <-chan R) func(...T) <-chan R {
	return func(data ...T) <-chan R {
		wg := new(sync.WaitGroup)
		merged := make(chan R)

		wg.Add(len(data))
		for _, item := range data {
			go func(c <-chan R) {
				defer wg.Done()
				for val := range c {
					merged <- val
				}
			}(fn(item))
		}

		go func() {
			wg.Wait()
			close(merged)
		}()

		return merged
	}
}
