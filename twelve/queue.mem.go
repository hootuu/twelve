package twelve

import (
	"github.com/hootuu/utils/errors"
	"sync"
)

type MemQueue struct {
	head *Job
	tail *Job
	db   map[string]*Job
	lock sync.RWMutex
}

func NewMemQueue() *MemQueue {
	return &MemQueue{
		head: nil,
		tail: nil,
		db:   make(map[string]*Job),
		lock: sync.RWMutex{},
	}
}

func (m *MemQueue) Head() (*Job, *errors.Error) {
	m.lock.RLock()
	defer m.lock.RUnlock()
	if m.head == nil {
		return nil, nil
	}
	return m.head, nil
}

func (m *MemQueue) Tail() (*Job, *errors.Error) {
	m.lock.RLock()
	defer m.lock.RUnlock()
	if m.tail == nil {
		return nil, nil
	}
	return m.tail, nil
}

func (m *MemQueue) Get(hash string) (*Job, *errors.Error) {
	m.lock.RLock()
	defer m.lock.RUnlock()
	j, ok := m.db[hash]
	if !ok {
		return nil, nil
	}
	return j, nil
}

func (m *MemQueue) Append(msg *Message) (*Job, *errors.Error) {
	hash := msg.ID
	m.lock.Lock()
	m.lock.Unlock()
	j, ok := m.db[hash]
	if ok {
		return nil, errors.Sys("repeated append: " + hash)
	}
	j = &Job{
		Hash:    hash,
		State:   Committed,
		Message: msg,
		Nxt:     "",
	}
	if m.head == nil {
		m.head = j
	}

	if m.tail == nil {
		m.tail = j
	} else {
		if m.tail.State != Confirmed {
			j.State = Pending
		}
		m.tail.Nxt = j.Hash
	}
	return j, nil
}

func (m *MemQueue) Remove(hash string) (*Job, *errors.Error) {
	m.lock.Lock()
	m.lock.Unlock()
	j, ok := m.db[hash]
	if !ok {
		return nil, nil
	}
	if len(j.Pre) != 0 {
		pre, ok := m.db[j.Pre]
		if ok {
			pre.Nxt = j.Nxt
		}
	}
	if len(j.Nxt) != 0 {
		nxt, ok := m.db[j.Nxt]
		if ok {
			nxt.Pre = j.Pre
		}
	}
	return nil, nil
}

func (m *MemQueue) Confirm(hash string) (*Job, *errors.Error) {
	m.lock.Lock()
	m.lock.Unlock()
	j, ok := m.db[hash]
	if !ok {
		return nil, errors.Sys("no such job: " + hash)
	}
	j.State = Confirmed
	return j, nil
}
