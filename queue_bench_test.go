package gocq

import (
	"runtime"
	"testing"
)

func BenchmarkQueue_Operations(b *testing.B) {
	cpus := uint(runtime.NumCPU())

	b.Run("Add", func(b *testing.B) {
		q := NewQueue(cpus, func(data int) int {
			return data * 2
		})
		defer q.Close()

		b.ResetTimer()
		for j := 0; j < b.N; j++ {
			<-q.Add(j)
		}
	})

	b.Run("AddAll", func(b *testing.B) {
		q := NewQueue(cpus, func(data int) int {
			return data * 2
		})
		defer q.Close()

		data := make([]int, b.N)
		for i := range data {
			data[i] = i
		}

		b.ResetTimer()
		out := q.AddAll(data...)
		for range out {
			// drain the channel
		}
	})

}

func BenchmarkPriorityQueue_Operations(b *testing.B) {
	cpus := uint(runtime.NumCPU())

	b.Run("Add", func(b *testing.B) {
		q := NewPriorityQueue(cpus, func(data int) int {
			return data * 2
		})
		defer q.Close()

		b.ResetTimer()
		for i := 0; i < b.N; i++ {
			out := q.Add(i, i%10)
			<-out
		}
	})

	b.Run("AddAll", func(b *testing.B) {
		q := NewPriorityQueue(cpus, func(data int) int {
			return data * 2
		})
		defer q.Close()

		data := make([]PQItem[int], b.N)
		for i := range data {
			data[i] = PQItem[int]{Value: i, Priority: i % 10}
		}

		b.ResetTimer()
		out := q.AddAll(data)

		for range out {
			// drain the channel
		}
	})

}
