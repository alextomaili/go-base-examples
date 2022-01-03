package contention

import (
	"strconv"
	"testing"
)

func BenchmarkRwStorage(b *testing.B) {
	for _, r := range []int{1, 5, 10, 20, 30, 40, 50, 100} {
		b.Run("RwStorage-w4-r"+strconv.Itoa(r), func(b *testing.B) {
			AggregateTest(b, NewRwStorage(), 4, r)
		})
	}
}

func BenchmarkSyncMapStorage(b *testing.B) {
	for _, r := range []int{1, 5, 10, 20, 30, 40, 50, 100} {
		b.Run("SyncMapStorage-w4-r"+strconv.Itoa(r), func(b *testing.B) {
			AggregateTest(b, NewSyncMapStorage(), 4, r)
		})
	}
}