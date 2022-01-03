package main

import (
	"fmt"
	"time"
)

func main() {
	c := make(chan bool)
	//var c chan bool

	go func() {
		time.Sleep(time.Hour)
		c <- true
	}()

	fmt.Println("Start & waint on chanell")

	b := <-c

	fmt.Println("received", b)
}
