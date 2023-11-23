package twelve

import (
	"fmt"
	"github.com/hootuu/utils/errors"
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
	counter   uint32
	done      chan ExpectState
	state     ExpectState
	timestamp time.Time
	replier   sync.Map
	lock      sync.Mutex
}

func ExpectID(peerID string, hash string, forArrow Arrow) string {
	return fmt.Sprintf("%s_%s_%d", peerID, hash, forArrow)
}

func NewExpect(id string, expect int) *Expect {
	e := &Expect{
		ID:        id,
		Expect:    uint32(expect),
		counter:   0,
		done:      make(chan ExpectState),
		state:     ExpectInit,
		timestamp: time.Now(),
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
	//if sys.RunMode.IsRd() {
	//	gLogger.Info("expect state change", zap.Int("target", int(s)),
	//		zap.String("id", e.ID))
	//}
}

func (e *Expect) Waiting(doFunc func() *errors.Error, onFunc func()) bool {
	//if sys.RunMode.IsRd() {
	//	fmt.Println("waiting for: ", zap.String("id", e.ID),
	//		zap.Uint32("expect", e.Expect))
	//}
	//go func() {
	//	e.SetState(ExpectWaiting)
	//	var newCounter = e.counter
	//	doneResult := ExpectInit
	//	for {
	//		end := false
	//		//if sys.RunMode.IsRd() {
	//		//	gLogger.Info("continue waiting for: ", zap.String("id", e.ID),
	//		//		zap.Uint32("expect", e.Expect), zap.Uint32("counter", newCounter))
	//		//}
	//		select {
	//		case <-e.waitGroup:
	//			newCounter = atomic.AddUint32(&e.counter, 1)
	//			//if sys.RunMode.IsRd() {
	//			//	gLogger.Info("done waiting for[get reply]: ", zap.String("id", e.ID),
	//			//		zap.Uint32("expect", e.Expect), zap.Uint32("counter", newCounter))
	//			//}
	//			if newCounter >= e.Expect {
	//				end = true
	//				doneResult = ExpectFinished
	//			}
	//		case result := <-e.done:
	//			end = true
	//			doneResult = result
	//		}
	//		if end {
	//			break
	//		}
	//	}
	//	e.confirmDone <- doneResult
	//}()
	err := doFunc()
	if err != nil {
		gLogger.Error("expect.doFunc failed",
			zap.String("id", e.ID),
			zap.Error(err))
		e.doCancel()
		return false
	}
	select {
	case result := <-e.done:
		fmt.Println("when result == E->", ExpectFinished, "<- vs ", result)
		success := false
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
		e.SetState(ExpectTimeout)
		return false
	}
}

func (e *Expect) Reply(peerID string) {
	//if sys.RunMode.IsRd() {
	//	fmt.Println("expect.Reply", zap.String("peerID", peerID),
	//		zap.String("id", e.ID),
	//		zap.Int("count", len(e.replier)))
	//}
	_, ok := e.replier.Load(peerID)
	if ok {
		gLogger.Warn("repeated peerID", zap.String("peerID", peerID))
		return
	}
	e.replier.Store(peerID, struct{}{})
	if e.IsEnd() {
		return
	}
	newCounter := atomic.AddUint32(&e.counter, 1)
	fmt.Println("newCounter==========>>>>>>>>>>>", newCounter)
	if newCounter == e.Expect {
		//if _, ok := <-e.done; !ok {
		//	fmt.Println("通道已关闭=================>>>>>>>>>>>>>>>>>>>")
		//} else {
		fmt.Println("GO FINISHED=================>>>>>>>>>>>>>>>>>>>")
		e.done <- ExpectFinished
		//}
	}
}

func (e *Expect) doCancel() {
	e.done <- ExpectCanceled
}

func (e *Expect) doClose() {
	//close(e.waitGroup)
	close(e.done)
	//close(e.confirmDone)
}

type ExpectFactory struct {
	db sync.Map
}

func NewExpectFactory() *ExpectFactory {
	return &ExpectFactory{}
}

func (ef *ExpectFactory) Build(peerID string, hash string, forArrow Arrow, expect int) *Expect {
	id := ExpectID(peerID, hash, forArrow)
	val, ok := ef.db.Load(id)
	if ok {
		fmt.Println(peerID, hash, forArrow, expect)
		sys.Exit(errors.Sys("repeated expect")) //todo
		if e, ok := val.(*Expect); ok {
			return e
		}
		return nil //todo
	}
	e := NewExpect(id, expect)
	ef.db.Store(id, e)
	return e
}

func (ef *ExpectFactory) MustGet(peerID string, hash string, forArrow Arrow) (*Expect, *errors.Error) {
	id := ExpectID(peerID, hash, forArrow)
	val, ok := ef.db.Load(id)
	if ok {
		if e, ok := val.(*Expect); ok {
			return e, nil
		}
		return nil, errors.Sys("invalid expect") //todo
	}
	return nil, errors.Sys("the expect not exists or has been destroyed")
}

func (ef *ExpectFactory) StartGC() {
	go func() {
		for {
			time.Sleep(60 * time.Second)

			var count int64 = 0
			var needRm []string
			ef.db.Range(func(key, value any) bool {
				count += 1
				id := key.(string)
				e := value.(*Expect)
				if e.IsEnd() {
					return true
				}
				needRm = append(needRm, id)
				return true
			})

			if len(needRm) > 0 {
				for _, id := range needRm {
					ef.db.Delete(id)
					if sys.RunMode.IsRd() {
						gLogger.Info("remove expect: ", zap.String("id", id))
					}
				}
			}

			fmt.Println("the EF length:==><==", count)
		}
	}()
}
