package contention

import (
	"strconv"
	"testing"
)

var readers = [...]int{1, 5, 10, 20, 30, 40, 50}

func BenchmarkRwStorage(b *testing.B) {
	for _, r := range readers {
		b.Run("RwStorage-w4-r"+strconv.Itoa(r), func(b *testing.B) {
			AggregateTest(b, NewRwStorage(), 4, r)
		})
	}
}

func BenchmarkRwSlStorage(b *testing.B) {
	for _, r := range readers {
		b.Run("RwSlStorage(1024)-w4-r"+strconv.Itoa(r), func(b *testing.B) {
			AggregateTest(b, NewRwSlStorage(10), 4, r)
		})
	}
}

func BenchmarkSyncMapStorage(b *testing.B) {
	for _, r := range readers {
		b.Run("SyncMapStorage-w4-r"+strconv.Itoa(r), func(b *testing.B) {
			AggregateTest(b, NewSyncMapStorage(), 4, r)
		})
	}
}
