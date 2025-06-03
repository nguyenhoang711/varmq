package main

import (
	"fmt"
	"time"

	"github.com/goptics/redisq"
	"github.com/goptics/varmq"
)

func main() {
	start := time.Now()
	defer func() {
		fmt.Println("Time taken:", time.Since(start))
	}()

	redisQueue := redisq.New("redis://localhost:6379")
	defer redisQueue.Close()
	rq := redisQueue.NewDistributedQueue("scraping_queue")
	defer rq.Close()
	defer rq.Listen()

	w := varmq.NewWorker(func(j varmq.Job[[]string]) {
		data := j.Data()
		url, id := data[0], data[1]
		fmt.Printf("Scraping url: %s, id: %s\n", url, id)
		// Simulate scraping work
		time.Sleep(2 * time.Second)
		fmt.Printf("Scraped url: %s, id: %s\n", url, id)
	}, 5)

	// Using redisq adapter (you can use any adapter that implements IDistributedQueue)
	q := w.WithDistributedQueue(rq)

	fmt.Println("pending jobs:", q.NumPending())
	fmt.Println("listening...")
}
