package contention

import (
	"sync"
	"sync/atomic"
	"time"
)

type (
	BIncrStorage struct {
		writers      int
		wc           int
		writeBatch   int32
		batches      [][2]map[Key]int64
		swapInterval time.Duration
		batchGen     int64
		counter      Counter
	}
)

func NewBIncrStorage(wc int, counter Counter) *BIncrStorage {
	r := &BIncrStorage{
		writers:      1,
		wc:           wc,
		writeBatch:   0,
		batches:      make([][2]map[Key]int64, 0, wc),
		swapInterval: time.Millisecond,
		counter:      counter,
	}
	for i := 0; i < wc; i++ {
		r.batches = append(r.batches, [2]map[Key]int64{make(map[Key]int64), make(map[Key]int64)})
	}
	go r.swap()
	return r
}

func (s *BIncrStorage) swap() {
	for {
		readBatch := atomic.LoadInt32(&s.writeBatch)
		atomic.StoreInt32(&s.writeBatch, (readBatch+1)&1)

		//за это время врайтеры гарантировано добегут
		time.Sleep(s.swapInterval)

		//apply batch to main storage
		for wn := 0; wn < s.wc; wn++ {
			for k, v := range s.batches[wn][readBatch] {
				m := Message{
					Key:   k,
					Value: v,
				}
				s.counter.Apply(m, wn)
				delete(s.batches[wn][readBatch], k)
			}
		}

		atomic.AddInt64(&s.batchGen, 1)
	}
}

func (s *BIncrStorage) BatchGeneration() int64 {
	return atomic.LoadInt64(&s.batchGen)
}

func (s *BIncrStorage) Consume(messages chan Message) {
	wg := sync.WaitGroup{}
	for i := 0; i < s.writers; i++ {
		wg.Add(1)
		go func(n int) {
			for m := range messages {
				s.Apply(m, n)
			}
			wg.Done()
		}(i)
	}
	wg.Wait()
}

//go:nosplit
func (s *BIncrStorage) Apply(msg Message, wn int) {
	k := msg.Key
	writeBatch := atomic.LoadInt32(&s.writeBatch)

	s.batches[wn][writeBatch][k] += msg.Value
}

func (s *BIncrStorage) Get(k Key) int64 {
	r := s.counter.Get(k)
	return r
}
