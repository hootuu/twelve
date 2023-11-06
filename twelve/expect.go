package twelve

import (
	"fmt"
	"sync"
)

type Expect struct {
	About   string `json:"about"`
	Expect  int    `json:"expect"`
	wg      sync.WaitGroup
	replier map[string]bool
	lock    sync.Mutex
}

func NewExpect(about string, expect int) *Expect {
	e := &Expect{
		About:   about,
		Expect:  expect,
		replier: make(map[string]bool),
	}
	e.wg.Add(e.Expect)
	return e
}

func (e *Expect) Waiting() {
	fmt.Println("waiting for: ", e.About, ", expect: ", e.Expect)
	e.wg.Wait()
}

func (e *Expect) Reply(peerID string) {
	e.lock.Lock()
	defer e.lock.Unlock()
	_, ok := e.replier[peerID]
	if ok {
		return
	}
	e.replier[peerID] = true
	e.wg.Done()
}

func (e *Expect) Cancel() {
	for i := 0; i < e.Expect; i++ {
		e.wg.Done()
	}
}
