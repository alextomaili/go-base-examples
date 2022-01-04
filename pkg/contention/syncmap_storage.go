package contention

import "sync"

type (
	SyncMapStorage struct {
		writers    int
		mutex      sync.RWMutex
		aggregates *sync.Map
	}
)

func NewSyncMapStorage() *SyncMapStorage {
	return &SyncMapStorage{
		writers:    4,
		mutex:      sync.RWMutex{},
		aggregates: &sync.Map{},
	}
}

func (s *SyncMapStorage) Consume(messages chan Message) {
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

func (s *SyncMapStorage) Apply(msg Message, _ int) {
	k := msg.Key

	s.mutex.Lock()
	actual, loaded := s.aggregates.LoadOrStore(k, msg.Value)
	if loaded {
		c := actual.(int64) + msg.Value
		s.aggregates.Store(k, c)
	}
	s.mutex.Unlock()
}

func (s *SyncMapStorage) Get(k Key) int64 {
	value, ok := s.aggregates.Load(k)
	if ok {
		return value.(int64)
	} else {
		return 0
	}
}
