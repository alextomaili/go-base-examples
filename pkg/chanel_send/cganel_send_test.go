package chanel_send

import (
	"testing"
)

func BenchmarkSendUnBufferedChanel(b *testing.B) {
	SendChanel(b, make(chan int64))
}

func BenchmarkSendBufferedChanel(b *testing.B) {
	SendChanel(b, make(chan int64, 1024))
}
