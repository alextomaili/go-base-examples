package contention

import "sync"

type (
	SyncMapStorage struct {
		mutex      sync.RWMutex
		aggregates *sync.Map
	}
)

func NewSyncMapStorage() *SyncMapStorage {
	return &SyncMapStorage{
		mutex:      sync.RWMutex{},
		aggregates: &sync.Map{},
	}
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
