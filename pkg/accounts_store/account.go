package accounts_store

import (
	"sync"
	"time"
)

type (
	Operation struct {
		id        uint64
		accountId uint64
		amount    float64
		opType    string
	}

	Account struct {
		id        uint64
		balance   float64
		opHistory []*Operation

		rwMutex sync.RWMutex
	}

	AccountSnapshot struct {
		AccountId uint64
		Balance   float64
		lastOps   []*Operation
		Timestamp time.Time
		LastOpId  uint64
	}
)

func NewOperation(id, accountId uint64, amount float64, opType string) *Operation {
	return &Operation{
		id:        id,
		accountId: accountId,
		amount:    amount,
		opType:    opType,
	}
}

func (op *Operation) GetId() uint64 {
	return op.id
}

func (op *Operation) GetAmount() float64 {
	return op.amount
}

func (sn *AccountSnapshot) GetLastOps() []*Operation {
	return sn.lastOps
}

func NewAccount(accountId uint64) *Account {
	return &Account{
		id:        accountId,
		opHistory: make([]*Operation, 0, 20),
	}
}

func (a *Account) Apply(op *Operation) bool {
	a.rwMutex.Lock()
	defer a.rwMutex.Unlock()

	if a.id != op.accountId {
		return false
	}

	a.opHistory = append(a.opHistory, op)
	a.balance = a.balance + op.amount

	return true
}

func (a *Account) Get(lastOpCount int) *AccountSnapshot {
	a.rwMutex.RLock()
	defer a.rwMutex.RUnlock()

	s := &AccountSnapshot{
		AccountId: a.id,
		Timestamp: time.Now(),
		Balance:   a.balance,
		lastOps:   make([]*Operation, 0, lastOpCount),
	}
	if len(a.opHistory) > 0 {
		s.LastOpId = a.opHistory[len(a.opHistory)-1].id
	}
	for i := len(a.opHistory) - 1; i >= 0 && len(s.lastOps) < lastOpCount; i-- {
		s.lastOps = append(s.lastOps, a.opHistory[i])
	}
	return s
}
