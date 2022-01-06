package contention

import (
	"sync"
	"sync/atomic"
	"time"
)

type (
	BIncrStorage struct {
		writers      int
		writeBatch   int32
		mCount       int32
		mLock        int32
		batches      [][2]map[Key]int64
		swapInterval time.Duration
		counter      Counter

		batchGen int64
	}
)

func NewBIncrStorage(wc int, swapInterval time.Duration, counter Counter) *BIncrStorage {
	r := &BIncrStorage{
		writers:      wc,
		writeBatch:   0,
		batches:      make([][2]map[Key]int64, 0, wc),
		swapInterval: swapInterval,
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
		time.Sleep(s.swapInterval)

		// raise flag
		atomic.StoreInt32(&s.mLock, 1)
		// wait mutators to be done
		for atomic.LoadInt32(&s.mCount) > 0 {
		}

		//swap batch
		readBatch := atomic.LoadInt32(&s.writeBatch)
		atomic.StoreInt32(&s.writeBatch, (readBatch+1)&1)

		//allow mutators
		atomic.StoreInt32(&s.mLock, 0)

		atomic.AddInt64(&s.batchGen, 1) //debug

		//apply batch to main storage
		for wn := 0; wn < s.writers; wn++ {
			for k, v := range s.batches[wn][readBatch] {
				m := Message{
					Key:   k,
					Value: v,
				}
				s.counter.Apply(m, wn)
				delete(s.batches[wn][readBatch], k)
			}
		}

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
	atomic.AddInt32(&s.mCount, 1)
	l := true
	for {
		mLock := atomic.LoadInt32(&s.mLock) //check lock
		if mLock == 1 {
			if l {
				atomic.AddInt32(&s.mCount, -1)
				l = false
			}
			continue
		}
		if !l {
			atomic.AddInt32(&s.mCount, 1)
		}
		break
	}

	k := msg.Key
	writeBatch := atomic.LoadInt32(&s.writeBatch)
	s.batches[wn][writeBatch][k] += msg.Value

	atomic.AddInt32(&s.mCount, -1)
}

func (s *BIncrStorage) Get(k Key) int64 {
	r := s.counter.Get(k)
	return r
}
