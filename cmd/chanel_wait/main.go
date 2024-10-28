package main

import (
	"fmt"
	"time"
)

func main() {
	c := make(chan bool)
	//var c chan bool

	go func() {
		time.Sleep(2 * time.Second)
		c <- true
	}()

	fmt.Println("Start & waint on chanell")

	b := <-c

	close(c)

	fmt.Println("received", b)
}
