package main

import (
	"fmt"
	"math/rand"
	"strconv"
	"time"

	"github.com/goptics/redisq"
	"github.com/goptics/varmq"
)

func generateJobID() string {
	const letters = "abcdefghijklmnopqrstuvwxyzABCDEFGHIJKLMNOPQRSTUVWXYZ0123456789"
	b := make([]byte, 8)
	for i := range b {
		b[i] = letters[rand.Intn(len(letters))]
	}
	return string(b)
}

func main() {
	start := time.Now()
	defer func() {
		fmt.Println("Time taken:", time.Since(start))
	}()

	redisQueue := redisq.New("redis://localhost:6379")
	defer redisQueue.Close()
	rq := redisQueue.NewDistributedQueue("scraping_queue")
	pq := varmq.NewDistributedQueue[[]string](rq)
	defer pq.Close()

	for i := range 1000 {
		id := generateJobID()
		data := []string{fmt.Sprintf("https://example.com/%s", strconv.Itoa(i)), id}
		pq.Add(data, varmq.WithJobId(id))
	}

	fmt.Println("added jobs")
	fmt.Println("pending jobs:", pq.NumPending())

	time.Sleep(5 * time.Second)
	pq.Purge()
	fmt.Println("purged jobs")
}
