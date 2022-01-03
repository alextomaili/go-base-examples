package main

import (
	"flag"
	"github.com/alextomaili/go-base-examples/pkg/chanel_send"
	"testing"
)

var (
	s = flag.Int("s", 0, "chanel size")
	n = flag.Int("n", 10000000, "counter")
)

func main() {
	flag.Parse()
	chanel_send.SendChanel(&testing.B{
		N: *n,
	}, make(chan int64, *s))
}
