package twelve

import (
	"fmt"
	"github.com/hootuu/utils/errors"
	"github.com/hootuu/utils/logger"
	"github.com/hootuu/utils/sys"
	"go.uber.org/zap"
	"sync"
	"sync/atomic"
	"time"
)

type ExpectState int

const (
	ExpectInit     ExpectState = 0
	ExpectWaiting  ExpectState = 1
	ExpectFinished ExpectState = 2
	ExpectTimeout  ExpectState = -1
	ExpectCanceled ExpectState = -2
)

const (
	DefaultExpectVoteCount   = 12
	DefaultExpectVoteCountRd = 12
	DefaultExpectTimeout     = 3600 * time.Second
)

type Expect struct {
	ID        string `json:"id"`
	Expect    uint32 `json:"expect"`
	waitGroup chan struct{}
	counter   uint32
	done      chan ExpectState
	state     ExpectState
	timestamp time.Time
	replier   map[string]struct{}
	lock      sync.Mutex
}

func ExpectID(peerID string, hash string, forType Type) string {
	return fmt.Sprintf("%s_%s_%d", peerID, hash, forType)
}

func NewExpect(id string, expect int) *Expect {
	e := &Expect{
		ID:        id,
		Expect:    uint32(expect),
		waitGroup: make(chan struct{}, 12*uint32(expect)),
		counter:   0,
		replier:   make(map[string]struct{}),
		done:      make(chan ExpectState),
		state:     ExpectInit,
	}
	return e
}

func (e *Expect) IsEnd() bool {
	switch e.state {
	case ExpectFinished, ExpectTimeout, ExpectCanceled:
		return true
	}
	return false
}

func (e *Expect) SetState(s ExpectState) {
	e.lock.Lock()
	defer e.lock.Unlock()
	e.state = s
	e.timestamp = time.Now()
	if sys.RunMode.IsRd() {
		gLogger.Info("expect state change", zap.Int("target", int(s)),
			zap.String("id", e.ID))
	}
}

func (e *Expect) Waiting(doFunc func() *errors.Error, onFunc func()) bool {
	if sys.RunMode.IsRd() {
		logger.Logger.Info("waiting for: ", zap.String("id", e.ID),
			zap.Uint32("expect", e.Expect))
	}
	go func() {
		e.SetState(ExpectWaiting)
		var newCounter = e.counter
		for {
			if sys.RunMode.IsRd() {
				logger.Logger.Info("continue waiting for: ", zap.String("id", e.ID),
					zap.Uint32("expect", e.Expect), zap.Uint32("counter", newCounter))
			}
			end := false
			select {
			case <-e.waitGroup:
				newCounter = atomic.AddUint32(&e.counter, 1)
				if sys.RunMode.IsRd() {
					logger.Logger.Info("done waiting for[get reply]: ", zap.String("id", e.ID),
						zap.Uint32("expect", e.Expect), zap.Uint32("counter", newCounter))
				}
			case result := <-e.done:
				end = true
				e.done <- result
				break
			}
			if end {
				break
			}
		}
		if newCounter >= e.Expect {
			e.done <- ExpectFinished
			close(e.done)
		}

	}()
	err := doFunc()
	if err != nil {
		gLogger.Error("expect.doFunc failed",
			zap.String("id", e.ID),
			zap.Error(err))
		e.doCancel()
	}
	select {
	case result := <-e.done:
		success := true
		switch result {
		case ExpectFinished:
			onFunc()
			e.SetState(ExpectFinished)
			success = true
		case ExpectCanceled:
			e.SetState(ExpectCanceled)
			success = false
		case ExpectTimeout:
			e.SetState(ExpectTimeout)
			success = false
		}
		return success
	case <-time.After(DefaultExpectTimeout):
		e.done <- ExpectTimeout
		e.SetState(ExpectTimeout)
		return false
	}
}

func (e *Expect) Reply(peerID string) {
	if sys.RunMode.IsRd() {
		logger.Logger.Info("expect.Reply", zap.String("peerID", peerID),
			zap.String("id", e.ID),
			zap.Int("count", len(e.replier)))
	}
	e.lock.Lock()
	defer e.lock.Unlock()
	_, ok := e.replier[peerID]
	if ok {
		logger.Logger.Warn("repeated peerID", zap.String("peerID", peerID))
		return
	}
	e.replier[peerID] = struct{}{}
	e.waitGroup <- struct{}{}
}

func (e *Expect) doCancel() {
	e.done <- ExpectCanceled
}

type ExpectFactory struct {
	db   map[string]*Expect
	lock sync.Mutex
}

func NewExpectFactory() *ExpectFactory {
	return &ExpectFactory{
		db:   make(map[string]*Expect),
		lock: sync.Mutex{},
	}
}

func (ef *ExpectFactory) Build(peerID string, hash string, forType Type, expect int) *Expect {
	ef.lock.Lock()
	defer ef.lock.Unlock()
	id := ExpectID(peerID, hash, forType)
	e, ok := ef.db[id]
	if ok {
		return e
	}
	e = NewExpect(id, expect)
	ef.db[id] = e
	return e
}

func (ef *ExpectFactory) MustGet(peerID string, hash string, forType Type) (*Expect, *errors.Error) {
	ef.lock.Lock()
	defer ef.lock.Unlock()
	id := ExpectID(peerID, hash, forType)
	e, ok := ef.db[id]
	if ok {
		return e, nil
	}
	return nil, errors.Sys("the expect not exists or has been destroyed")
}

func (ef *ExpectFactory) StartGC() {
	go func() {
		for {
			time.Sleep(5 * time.Second)
			var needRm []string
			for _, e := range ef.db {
				if !e.IsEnd() {
					continue
				}
				if time.Now().Sub(e.timestamp) < 10*time.Second {
					continue
				}
				needRm = append(needRm, e.ID)
			}
			if len(needRm) > 0 {
				ef.lock.Lock()
				for _, id := range needRm {
					delete(ef.db, id)
					if sys.RunMode.IsRd() {
						gLogger.Info("remove expect: ", zap.String("id", id))
					}
				}
				ef.lock.Unlock()
			}
		}
	}()
}
