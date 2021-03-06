package contention

import "sync"

type (
	RwSlStorage struct {
		writers    int
		bn         int
		mutex      []sync.RWMutex
		aggregates []map[Key]int64
	}
)

func NewRwSlStorage(bnPov2 int, wc int) *RwSlStorage {
	bn := 2 << (bnPov2 - 1)
	r := &RwSlStorage{
		writers:    wc,
		bn:         bn,
		mutex:      make([]sync.RWMutex, 0, bn),
		aggregates: make([]map[Key]int64, 0, bn),
	}
	for i := 0; i < bn; i++ {
		r.mutex = append(r.mutex, sync.RWMutex{})
		r.aggregates = append(r.aggregates, make(map[Key]int64))
	}
	return r
}

func (s *RwSlStorage) hashSlot(k Key) int {
	h := k.x<<16 + k.y
	h = h ^ (h >> 16) // from java.util.HashMap, java 1.8
	return h & (s.bn - 1)
}

func (s *RwSlStorage) Consume(messages chan Message) {
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

func (s *RwSlStorage) Apply(msg Message, _ int) {
	k := msg.Key
	i := s.hashSlot(k)

	s.mutex[i].Lock()
	s.aggregates[i][k] += msg.Value
	s.mutex[i].Unlock()
}

func (s *RwSlStorage) Get(k Key) int64 {
	i := s.hashSlot(k)

	s.mutex[i].RLock()
	r := s.aggregates[i][k]
	s.mutex[i].RUnlock()
	return r
}
