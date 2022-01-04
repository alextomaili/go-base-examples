package contention

import "sync"

type (
	RwStorage struct {
		writers    int
		mutex      sync.RWMutex
		aggregates map[Key]int64
	}
)

func NewRwStorage() *RwStorage {
	return &RwStorage{
		writers:    4,
		mutex:      sync.RWMutex{},
		aggregates: make(map[Key]int64),
	}
}

func (s *RwStorage) Consume(messages chan Message) {
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
