package contention

import (
	"sync"
	"sync/atomic"
	"time"
)

type (
	BIncrStorage struct {
		writers        int
		writeBatch     int32
		pendingWriters int32
		swapLock       int32
		batches        [][2]map[Key]int64
		swapInterval   time.Duration

		storage Counter

		batchGen int64 //debug
	}
)

func NewBIncrStorage(wc int, swapInterval time.Duration, counter Counter) *BIncrStorage {
	r := &BIncrStorage{
		writers:      wc,
		writeBatch:   0,
		batches:      make([][2]map[Key]int64, 0, wc),
		swapInterval: swapInterval,
		storage:      counter,
	}
	for i := 0; i < wc; i++ {
		r.batches = append(r.batches, [2]map[Key]int64{make(map[Key]int64), make(map[Key]int64)})
	}
	go r.swapAndApplyBatch()
	return r
}

func (s *BIncrStorage) applyBatchToStorage(readBatch int32) {
	for wn := 0; wn < s.writers; wn++ {
		for k, v := range s.batches[wn][readBatch] {
			m := Message{
				Key:   k,
				Value: v,
			}
			s.storage.Apply(m, wn)
			delete(s.batches[wn][readBatch], k)
		}
	}
}

func (s *BIncrStorage) swapAndApplyBatch() {
	for {
		time.Sleep(s.swapInterval)

		atomic.StoreInt32(&s.swapLock, 1)
		// wait for all pending readers
		for {
			if atomic.LoadInt32(&s.pendingWriters) == 0 {
				break
			}
		}

		//swap batch
		readBatch := atomic.LoadInt32(&s.writeBatch)
		atomic.StoreInt32(&s.writeBatch, (readBatch+1)&1)

		atomic.StoreInt32(&s.swapLock, 0)

		atomic.AddInt64(&s.batchGen, 1) //debug

		//apply batch to main storage
		s.applyBatchToStorage(readBatch)
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
	atomic.AddInt32(&s.pendingWriters, 1)

	holdLock := true
	for {
		if atomic.LoadInt32(&s.swapLock) == 0 {
			break
		}
		if holdLock {
			holdLock = false
			atomic.AddInt32(&s.pendingWriters, -1)
		}
	}
	// <<<
	if !holdLock {
		atomic.AddInt32(&s.pendingWriters, 1)
	}


	writeBatch := atomic.LoadInt32(&s.writeBatch)
	s.batches[wn][writeBatch][msg.Key] += msg.Value

	atomic.AddInt32(&s.pendingWriters, -1)
}

func (s *BIncrStorage) Get(k Key) int64 {
	r := s.storage.Get(k)
	return r
}
