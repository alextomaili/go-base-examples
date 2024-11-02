package main

import (
	"fmt"
	"time"
)

func main() {
	//send2c()
	//sendWithWait()
	sendNonBlock()
}

type P struct {
	a int64
	s string
	b bool
	m [2]uint64
}

func send2c() {
	c := make(chan *P)

	go func() {
		time.Sleep(2 * time.Second)
		p := &P{
			a: 645,
			s: "payload",
		}
		c <- p
	}()

	fmt.Println("Start & waint on chanell")
	p := <-c
	close(c)

	fmt.Println("received", p)
}

func sendWithWait() {
	g := 1
	c := make(chan *P)

	go func(g int) {
		c <- &P{
			a: 645,
			s: "payload",
		}
		fmt.Printf("[%d] %v: Was sent\n", g, time.Now())
	}(2)

	fmt.Printf("[%d] %v: Sleep 2s\n", g, time.Now())
	time.Sleep(2 * time.Second)

	fmt.Printf("[%d] %v: Start & waits on chanel\n", g, time.Now())
	p := <-c
	close(c)

	fmt.Printf("[%d] %v: received: %v\n", g, time.Now(), p)
}

func sendNonBlock() {
	g := 1
	c := make(chan *P)

	go func(g int) {
		p := &P{
			a: 645,
			s: "payload",
		}

		select {
		case c <- p:
			fmt.Printf("[%d] %v: Was sent now\n", g, time.Now())
		default:
			fmt.Printf("[%d] %v: Try send latter\n", g, time.Now())
			go func(g int) {
				c <- p
				fmt.Printf("[%d] %v: Was sent async\n", g, time.Now())
			}(3)
		}

		fmt.Printf("[%d] %v: Quit \n", g, time.Now())
	}(2)

	fmt.Printf("[%d] %v: Sleep 2s\n", g, time.Now())
	time.Sleep(2 * time.Second)

	fmt.Printf("[%d] %v: Start & waits on chanel\n", g, time.Now())
	t := time.After(5 * time.Second)
	q := false
	for !q {
		select {
		case <-t:
			//q = true
			close(c)
		case p, f := <-c:
			fmt.Printf("[%d] %v: received: %v %v\n", g, time.Now(), p, f)
			if !f {
				q = true
			}
		default:
			fmt.Printf("[%d] %v: Wait for data ....\n", g, time.Now())
			time.Sleep(time.Second)
		}
	}

	//close(c)
	time.Sleep(time.Second)
	fmt.Printf("[%d] %v: Quit \n", g, time.Now())
}
