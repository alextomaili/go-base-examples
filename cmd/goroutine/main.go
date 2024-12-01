package main

import (
	"fmt"
	"runtime"
	"sync"
	"sync/atomic"
	"time"
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
func (w *MyWg2) Wait(timeout time.Duration) bool {
	t := time.NewTimer(timeout)
	defer t.Stop()

	counter, isTimeout := w.readChanel(t)
	if isTimeout {
		return false
	}

	for counter != 0 {
		c, isTimeout := w.readChanel(t)
		if isTimeout {
			return false
		}

		counter = counter + c
	}

	fmt.Printf("Wg Done\n")
	return true
}

func (w *MyWg2) readChanel(t *time.Timer) (int64, bool) {
	select {
	case cv := <-w.c:
		return int64(cv), false
	case <-t.C:
		return 0, true
	}
}

func main() {
	runtime.GOMAXPROCS(1)

	n := 10
	wg := NewMyWg2(1024)

	for i := 0; i < n; i++ {
		wg.Add(1)
		go func(n int) {
			job(n)
			time.Sleep(time.Hour)
			wg.Add(-1)
		}(i)
	}

	if wg.Wait(time.Second) {
		fmt.Printf("Wait Dome\n")
	} else {
		fmt.Printf("Wait Timeout\n")
	}

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
