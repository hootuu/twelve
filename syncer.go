package twelve

import (
	"fmt"
	"github.com/hootuu/tome/kt"
	"github.com/hootuu/tome/vn"
	"github.com/hootuu/utils/errors"
	"go.uber.org/zap"
	"time"
)

const (
	syncIntervalTime = 30 * time.Second
)

type SyncDoCheckFunc func(letter *Letter) bool
type SyncDoSyncFunc func(letter *Letter) *errors.Error
type SyncDoNotifyFunc func() *errors.Error

type Syncer struct {
	vnID          vn.ID
	chain         kt.Chain
	doCheck       SyncDoCheckFunc
	doSync        SyncDoSyncFunc
	doNotify      SyncDoNotifyFunc
	synced        bool
	lstSyncedTime time.Time
}

func NewSyncer(
	vnID vn.ID,
	chain kt.Chain,
	doCheckFunc SyncDoCheckFunc,
	doSyncFunc SyncDoSyncFunc,
	doNotifyFunc SyncDoNotifyFunc,
) *Syncer {
	return &Syncer{
		vnID:          vnID,
		chain:         chain,
		doCheck:       doCheckFunc,
		doSync:        doSyncFunc,
		doNotify:      doNotifyFunc,
		synced:        false,
		lstSyncedTime: time.Time{},
	}
}

func (sync *Syncer) Startup() {
	timer := time.NewTicker(syncIntervalTime)
	for {
		select {
		case <-timer.C:
			fmt.Println("执行同步消息发送") // todo del
			err := sync.doNotify()
			if err != nil {
				gLogger.Error("Syncer.doNotify() failed [ignore]", zap.String("VN", sync.vnID.S()),
					zap.String("Chain", sync.chain.S()))
			}
		}
	}
}

func (sync *Syncer) On(letter *Letter) *errors.Error {
	if letter.Arrow != InvariableArrow {
		return errors.Assert(InvariableArrow.S(), letter.Arrow.S())
	}
	if letter.Vn != sync.vnID {
		return errors.Assert(sync.vnID.S(), letter.Vn.S())
	}
	if letter.Chain != sync.chain {
		return errors.Assert(sync.chain.S(), letter.Chain.S())
	}
	bSynced := sync.doCheck(letter)
	if bSynced {
		return nil
	}
	//todo wait and do sync
	expect, isNew, err := gExpectFactory.GetOrBuild("xx", letter.Invariable.S(), InvariableArrow, 3)
	if err != nil {
		return err
	}
	if isNew {
		go expect.Waiting(func() *errors.Error {
			return nil
		}, func() {
			sync.doSync(letter)
		})
	} else {
		expect.Reply(letter.From.S())
	}
	return nil
}
