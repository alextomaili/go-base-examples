package contention

import (
	"math/rand"
	"sync"
	"testing"
)

type (
	Key struct {
		x, y int
	}
	Message struct {
		Key   Key
		Value int64
	}
	Counter interface {
		Apply(msg Message)
		Get(key Key) int64
	}
)

var result []int64

func fillStorage(storage Counter, k []Key, v []int64) {
	for i, k := range k {
		m := Message{
			Key: k, Value: v[i],
		}
		storage.Apply(m)
	}
}

func AggregateTest(b *testing.B, storage Counter, writers, readers int, waitReaders bool) {
	b.StopTimer()

	keys := make([]Key, b.N)
	values := make([]int64, b.N)
	for i := 0; i < b.N; i++ {
		keys = append(keys, Key{x: rand.Intn(10000), y: rand.Intn(255)})
		values = append(values, rand.Int63())
	}
	//все ключи будут присутствовать в тесте заранее
	fillStorage(storage, keys, values)

	start := sync.WaitGroup{}
	start.Add(1)

	wg := sync.WaitGroup{}
	for i := 0; i < writers; i++ {
		wg.Add(1)
		go func() {
			start.Wait()
			fillStorage(storage, keys, values)
			wg.Done()
		}()
	}

	result = make([]int64, readers)
	for i := 0; i < readers; i++ {
		if waitReaders {
			wg.Add(1)
		}
		go func(n int) {
			start.Wait()
			for _, k := range keys {
				result[n] += storage.Get(k)
			}
			if waitReaders {
				wg.Done()
			}
		}(i)
	}

	b.StartTimer()
	start.Done()
	wg.Wait()
}
