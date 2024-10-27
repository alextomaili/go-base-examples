package main

import (
	"fmt"
	"runtime"
	"sync"
)

/*
	maxProcs := runtime.NumCPU()
	runtime.GOMAXPROCS(maxProcs)
*/

func main() {
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
			wg.Done()
		}(i)
	}

	wg.Wait()

	fmt.Printf("All jobs done\n")
}

func job(n int) {
	fmt.Printf("Hi i am job # %d\n", n)
}
