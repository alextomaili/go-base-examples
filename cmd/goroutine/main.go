package main

import (
	"fmt"
	"runtime"
	"sync"
	"sync/atomic"
)

/*
	maxProcs := runtime.NumCPU()
	runtime.GOMAXPROCS(maxProcs)
*/

type MyWg struct {
	counter int64
	bufSize int
	c       chan int
}

func NewMyWg(bufSize int) *MyWg {
	return &MyWg{
		bufSize: bufSize,
		c:       make(chan int, bufSize),
	}
}

// Add change [WaitGroup] counter.
func (w *MyWg) Add(delta int) {
	newC := atomic.AddInt64(&w.counter, int64(delta))
	if newC >= int64(w.bufSize) {
		panic("WaitGroup buffer size excited")
	} else if newC < 0 {
		panic("negative WaitGroup counter")
	}
	w.c <- 1
}

// Wait blocks until the [WaitGroup] counter is zero.
func (w *MyWg) Wait() {
	for atomic.LoadInt64(&w.counter) != 0 {
		<-w.c
	}
}

type MyWg2 struct {
	c chan int
}

func NewMyWg2(bufSize int) *MyWg2 {
	return &MyWg2{
		c: make(chan int, bufSize),
	}
}

// Add change [WaitGroup] counter.
func (w *MyWg2) Add(delta int) {
	w.c <- delta
}

// Wait blocks until the [WaitGroup] counter is zero.
func (w *MyWg2) Wait() {
	counter := int64(<-w.c)
	for counter != 0 {
		counter = counter + int64(<-w.c)
	}
	fmt.Printf("Done")
}

func main() {
	n := 10
	wg := NewMyWg2(1024)

	for i := 0; i < n; i++ {
		wg.Add(1)
		go func(n int) {
			job(n)
			wg.Add(-1)
		}(i)
	}

	wg.Wait()

	fmt.Printf("All jobs done\n")
}

func main_() {
	var (
		n  int = 10
		wg sync.WaitGroup
	)
	fmt.Printf("\n\nStarting %d goroutines\n", n)

	fmt.Printf("i have: %d CPU\n", runtime.NumCPU())
	pmp := runtime.GOMAXPROCS(1)
	fmt.Printf("previous i had: %d CPU\n", pmp)

	for i := 0; i < n; i++ {
		wg.Add(1)
		go func(n int) {
			job(n)
			wg.Add(-1)
		}(i)
	}

	wg.Wait()

	fmt.Printf("All jobs done\n")
}

func job(n int) {
	fmt.Printf("Hi i am job # %d\n", n)
}
