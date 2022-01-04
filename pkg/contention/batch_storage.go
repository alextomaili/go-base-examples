package contention

import (
	"sync"
	"time"
)

type (
	BatchStorage struct {
		writers      int
		writeBatch   int32
		swapInterval time.Duration
		counter      Counter
	}
)

func NewBatchStorage(wc int, counter Counter) *BatchStorage {
	r := &BatchStorage{
		writers:      wc,
		swapInterval: time.Millisecond,
		counter:      counter,
	}
	return r
}

func (s *BatchStorage) applyBatch(batch map[Key]int64, wn int) {
	for k, v := range batch {
		m := Message{
			Key:   k,
			Value: v,
		}
		s.Apply(m, wn)
		delete(batch, k)
	}
}

func (s *BatchStorage) worker(messages chan Message, chA, chB chan struct{}, active bool, wn int) {
	batch := make(map[Key]int64, 100000)
	var t *time.Timer

	for {
		if active {
			if t == nil {
				t = time.NewTimer(s.swapInterval)
			} else {
				t.Reset(s.swapInterval)
			}
			for {
				select {
				case <-t.C:
					select {
					case chB <- struct{}{}:
						s.applyBatch(batch, wn)
						active = false
						//batch = make(map[Key]int64)
						break
					default:
						t.Reset(s.swapInterval)
					}
				case m, ok := <-messages:
					if !ok {
						close(chB)
						s.applyBatch(batch, wn)
						return
					}
					batch[m.Key] += m.Value
				}
			}
		} else {
			select {
			case _, ok := <-chA:
				if !ok {
					close(chB)
					s.applyBatch(batch, wn)
					return
				}
				active = true
			}
		}
	}
}

func (s *BatchStorage) Consume(messages chan Message) {
	wg := sync.WaitGroup{}
	for i := 0; i < s.writers; i++ {
		wg.Add(2)
		chA := make(chan struct{})
		chB := make(chan struct{})
		go func(n int) {
			s.worker(messages, chA, chB, true, n)
			wg.Done()
		}(i)
		go func(n int) {
			s.worker(messages, chB, chA, false, n)
			wg.Done()
		}(i)
	}
	wg.Wait()
}

func (s *BatchStorage) Apply(msg Message, wn int) {
	s.counter.Apply(msg, wn)
}

func (s *BatchStorage) Get(k Key) int64 {
	r := s.counter.Get(k)
	return r
}
