package contention

import "sync"

type (
	RwStorage struct {
		mutex      sync.RWMutex
		aggregates map[Key]int64
	}
)

func NewRwStorage() *RwStorage {
	return &RwStorage{
		mutex:      sync.RWMutex{},
		aggregates: make(map[Key]int64),
	}
}

func (s *RwStorage) Apply(msg Message, _ int) {
	k := msg.Key
	s.mutex.Lock()
	s.aggregates[k] += msg.Value
	s.mutex.Unlock()
}

func (s *RwStorage) Get(k Key) int64 {
	s.mutex.RLock()
	r := s.aggregates[k]
	s.mutex.RUnlock()
	return r
}
