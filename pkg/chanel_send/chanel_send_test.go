package chanel_send

import (
	"testing"
)

func BenchmarkSendChanel(b *testing.B) {
	b.Run("BenchmarkSendUnBufferedChanel", func(b *testing.B) {
		SendChanel(b, make(chan int64))
	})
	b.Run("BenchmarkSendBufferedChanel", func(b *testing.B) {
		SendChanel(b, make(chan int64, 1024))
	})
}
