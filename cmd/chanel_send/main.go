package main

import (
	"flag"
	"github.com/alextomaili/go-base-examples/pkg/chanel_send"
	"sync"
	"testing"
)

var (
	s = flag.Int("s", 0, "chanel size")
	n = flag.Int("n", 10000000, "counter")
)

func main() {
	m := sync.Mutex{}
	m.Lock()
	m.Lock()

	flag.Parse()
	chanel_send.SendChanel(&testing.B{
		N: *n,
	}, make(chan int64, *s))
}
