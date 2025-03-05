package main

import (
	"fmt"
	"time"

	"github.com/fahimfaisaal/gocq/queue"
)

func main() {
	start := time.Now()
	defer func() {
		fmt.Printf("Took %s\n", time.Since(start))
	}()
	q := queue.NewPQ(10, func(data int) int {
		fmt.Printf("Started Worker %d\n", data)
		for i := 0; i < 1e10; i++ {
			// do nothing
		}
		fmt.Printf("Ended Worker %d\n", data)
		return data
	})
	defer q.WaitAndClose()

	vals := make([]int, 100)
	for i := 0; i < 100; i++ {
		vals[i] = i + 1
	}

	q.AddAll(1, vals...)
	fmt.Println("All tasks have been added")

	time.Sleep(time.Second * 5)
	q.Close()
	fmt.Println("Closed the queue successfully")
	q.Init()
	fmt.Println("Re-initial queue")

	time.Sleep(time.Second * 5)

	for i := 0; i < 30; i++ {
		q.Add(i+1, 1)
	}
	fmt.Println("All tasks have been added 2")
}
