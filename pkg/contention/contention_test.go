package contention

import (
	"strconv"
	"testing"
	"time"
)

var writers = 4
var readers = [...]int{0, 1, 5, 10, 20, 30, 40, 50}
var waitReaders = false
var swapInterval = time.Microsecond * 100

func BenchmarkRwStorage(b *testing.B) {
	b.Run("RwStorage-w1-r0", func(b *testing.B) {
		AggregateTest(b, NewRwStorage(1), 1, 0, waitReaders)
	})
	for _, r := range readers {
		b.Run("RwStorage-w4-r"+strconv.Itoa(r), func(b *testing.B) {
			AggregateTest(b, NewRwStorage(writers), writers, r, waitReaders)
		})
	}
}

func BenchmarkRwSlStorage(b *testing.B) {
	b.Run("RwSlStorage(1024)-w1-r0", func(b *testing.B) {
		AggregateTest(b, NewRwSlStorage(10, 1), 1, 0, waitReaders)
	})
	for _, r := range readers {
		b.Run("RwSlStorage(1024)-w4-r"+strconv.Itoa(r), func(b *testing.B) {
			AggregateTest(b, NewRwSlStorage(10, writers), writers, r, waitReaders)
		})
	}
}

func BenchmarkSyncMapStorage(b *testing.B) {
	b.Run("SyncMapStorage-w1-r0", func(b *testing.B) {
		AggregateTest(b, NewSyncMapStorage(1), 1, 0, waitReaders)
	})
	for _, r := range readers {
		b.Run("SyncMapStorage-w4-r"+strconv.Itoa(r), func(b *testing.B) {
			AggregateTest(b, NewSyncMapStorage(writers), writers, r, waitReaders)
		})
	}
}

func BenchmarkBatchStorage(b *testing.B) {
	b.Run("BatchStorage-w1-r0", func(b *testing.B) {
		storage := NewBatchStorage(1, swapInterval, NewRwSlStorage(10, 1))
		AggregateTest(b, storage, 1, 0, waitReaders)
	})
	for _, r := range readers {
		b.Run("BatchStorage-w4-r"+strconv.Itoa(r), func(b *testing.B) {
			storage := NewBatchStorage(writers, swapInterval, NewRwSlStorage(10, writers))
			AggregateTest(b, storage, writers, r, waitReaders)
		})
	}
}

func BenchmarkBIncrStorage(b *testing.B) {
	b.Run("BIncrStorage-w1-r0", func(b *testing.B) {
		storage := NewBIncrStorage(1, swapInterval, NewRwSlStorage(10, 1))
		AggregateTest(b, storage, 1, 0, waitReaders)
		//b.Logf("batch generation %v", storage.BatchGeneration())
	})
	for _, r := range readers {
		b.Run("BIncrStorage-w4-r"+strconv.Itoa(r), func(b *testing.B) {
			storage := NewBIncrStorage(writers, swapInterval, NewRwSlStorage(10, writers))
			AggregateTest(b, storage, writers, r, waitReaders)
			//b.Logf("batch generation %v", storage.BatchGeneration())
		})
	}
}
