package contention

import (
	"strconv"
	"testing"
)

var readers = [...]int{0, 1, 5, 10, 20, 30, 40, 50}
var waitReaders = false

func BenchmarkRwStorage(b *testing.B) {
	b.Run("RwStorage-w1-r0", func(b *testing.B) {
		AggregateTest(b, NewRwStorage(), 1, 0, waitReaders)
	})
	for _, r := range readers {
		b.Run("RwStorage-w4-r"+strconv.Itoa(r), func(b *testing.B) {
			AggregateTest(b, NewRwStorage(), 4, r, waitReaders)
		})
	}
}

func BenchmarkRwSlStorage(b *testing.B) {
	b.Run("RwSlStorage(1024)-w1-r0", func(b *testing.B) {
		AggregateTest(b, NewRwSlStorage(10), 1, 0, waitReaders)
	})
	for _, r := range readers {
		b.Run("RwSlStorage(1024)-w4-r"+strconv.Itoa(r), func(b *testing.B) {
			AggregateTest(b, NewRwSlStorage(10), 4, r, waitReaders)
		})
	}
}

func BenchmarkSyncMapStorage(b *testing.B) {
	b.Run("SyncMapStorage-w1-r0", func(b *testing.B) {
		AggregateTest(b, NewSyncMapStorage(), 1, 0, waitReaders)
	})
	for _, r := range readers {
		b.Run("SyncMapStorage-w4-r"+strconv.Itoa(r), func(b *testing.B) {
			AggregateTest(b, NewSyncMapStorage(), 4, r, waitReaders)
		})
	}
}
