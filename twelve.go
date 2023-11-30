package twelve

import (
	"github.com/hootuu/tome/kt"
	"github.com/hootuu/tome/nd"
	"github.com/hootuu/tome/vn"
	"github.com/hootuu/utils/errors"
	"github.com/hootuu/utils/sys"
	"github.com/hootuu/utils/types/pagination"
	"go.uber.org/zap"
	"sync"
)

type Option struct {
	Expect int `json:"expect"`
}

type Twelve struct {
	vnID     vn.ID
	chain    kt.Chain
	queue    *Queue
	notifier *Notifier
	node     *nd.Lead
	here     *nd.Node
	option   *Option
	lock     sync.Mutex
}

func NewTwelve(
	vnID vn.ID,
	chain kt.Chain,
	twelveNode ITwelveNode,
	option *Option,
) (*Twelve, *errors.Error) {
	if err := chain.Verify(); err != nil {
		return nil, err
	}
	if twelveNode == nil {
		return nil, errors.Verify("require twelve node")
	}
	if option == nil {
		if sys.RunMode.IsRd() {
			option = &Option{Expect: DefaultExpectVoteCountRd}
		} else {
			option = &Option{Expect: DefaultExpectVoteCount}
		}
	}
	tw := &Twelve{
		chain:    chain,
		notifier: NewNotifier(twelveNode),
		node:     twelveNode.Node().Lead(),
		option:   option,
	}
	var err *errors.Error
	tw.queue, err = NewQueue(vnID, chain)
	if err != nil {
		return nil, err
	}
	tw.notifier.BindListener(tw)
	tw.here = twelveNode.Node()
	return tw, nil
}

func (tw *Twelve) OnRequest(letter *Letter) {
	tx, err := tw.queue.Append(letter)
	if err != nil {
		gLogger.Error("twelve.queue.Append(msg) failed",
			zap.Any("hash", letter), zap.Error(err))
		return
	}
	err = tw.doPrepare(tx.Hash)
	if err != nil {
		gLogger.Error("tw.doPrepare(j.Hash) failed",
			zap.String("tx", tx.Hash), zap.Error(err))
		return
	}
}

func (tw *Twelve) doPrepare(hash string) *errors.Error {
	if sys.RunMode.IsRd() {
		gLogger.Info("do prepare", zap.String("hash", hash))
	}
	tx, err := tw.queue.MustGet(hash)
	if err != nil {
		return err
	}
	prepareLetter := NewLetter(tw.vnID, tw.chain, tx.Letter.Invariable, PrepareArrow, tw.here.ID)
	err = prepareLetter.Sign(tw.here.PRI)
	if err != nil {
		return err
	}
	expect := gExpectFactory.Build(tw.here.ID.S(), tx.Hash, PrepareArrow, tw.option.Expect)
	if sys.RunMode.IsRd() {
		gLogger.Info("build expect for prepare", zap.String("peerID", tw.here.ID.S()),
			zap.String("hash", hash),
			zap.Int32("arrow", int32(PrepareArrow)),
			zap.String("letter.invariable", hash))
	}
	go expect.Waiting(func() *errors.Error {
		return tw.notifier.Notify(prepareLetter)
	}, func() {
		err := tw.doCommit(tx.Hash)
		if err != nil {
			gLogger.Error("doCommit failed", zap.String("tx", tx.Hash), zap.Error(err))
		}
	})
	return nil
}

func (tw *Twelve) doCommit(hash string) *errors.Error {
	if sys.RunMode.IsRd() {
		gLogger.Info("do commit", zap.String("hash", hash))
	}
	tx, err := tw.queue.MustGet(hash)
	if err != nil {
		return err
	}
	committedLetter := NewLetter(tw.vnID, tw.chain, tx.Letter.Invariable, CommittedArrow, tw.here.ID)
	err = committedLetter.Sign(tw.here.PRI)
	if err != nil {
		return err
	}
	expect := gExpectFactory.Build(tw.node.ID.S(), tx.Hash, CommittedArrow, tw.option.Expect)
	if sys.RunMode.IsRd() {
		gLogger.Info("build expect for commit", zap.String("peerID", tw.here.ID.S()),
			zap.String("hash", hash),
			zap.Int32("arrow", int32(CommittedArrow)),
			zap.String("letter.invariable", hash))
	}
	go expect.Waiting(func() *errors.Error {
		return tw.notifier.Notify(committedLetter)
	}, func() {
		err := tw.doConfirm(tx.Hash)
		if err != nil {
			gLogger.Error("doConfirm failed", zap.String("tx", tx.Hash), zap.Error(err))
		}
	})
	return nil
}

func (tw *Twelve) doConfirm(hash string) *errors.Error {
	if sys.RunMode.IsRd() {
		gLogger.Info("do confirm", zap.String("hash", hash))
	}
	tx, err := tw.queue.MustGet(hash)
	if err != nil {
		return err
	}

	confirmedLetter := NewLetter(tw.vnID, tw.chain, tx.Letter.Invariable, ConfirmedArrow, tw.here.ID)
	err = confirmedLetter.Sign(tw.here.PRI)
	if err != nil {
		return err
	}
	expect := gExpectFactory.Build(tw.node.ID.S(), tx.Hash, ConfirmedArrow, tw.option.Expect)
	if sys.RunMode.IsRd() {
		gLogger.Info("build expect for confirm", zap.String("peerID", tw.here.ID.S()),
			zap.String("hash", hash),
			zap.Int32("arrow", int32(ConfirmedArrow)),
			zap.String("letter.invariable", hash))
	}
	go expect.Waiting(func() *errors.Error {
		return tw.notifier.Notify(confirmedLetter)
	}, func() {
		_, err := tw.queue.Confirm(tx.Hash)
		if err != nil {
			gLogger.Error("tw.queue.Confirm(tx.Hash) failed", zap.String("hash", tx.Hash))
		}
	})
	return nil
}

func (tw *Twelve) doOnMessage(letter *Letter) {
	hash := letter.Invariable.S()
	expect, err := gExpectFactory.MustGet(tw.here.ID.S(), hash, letter.Arrow)
	if err != nil {
		gLogger.Error("no such expect",
			zap.String("peerID", tw.here.ID.S()),
			zap.String("hash", hash),
			zap.Int32("arrow", int32(letter.Arrow)),
			zap.String("letter.invariable", hash))
		return
	}
	go expect.Reply(letter.From.S())
}

func (tw *Twelve) OnPrepare(letter *Letter) {
	tw.doOnMessage(letter)
}

func (tw *Twelve) OnCommitted(letter *Letter) {
	tw.doOnMessage(letter)
}

func (tw *Twelve) OnConfirmed(letter *Letter) {
	tw.doOnMessage(letter)
}

func (tw *Twelve) Start() {
	tw.notifier.Listening()
}

func (tw *Twelve) Tail() (*Tx, *errors.Error) {
	return tw.queue.Tail()
}

func (tw *Twelve) List(lstHash string, limit pagination.Limit) ([]*Tx, string, *errors.Error) {
	return tw.queue.List(lstHash, limit)
}

func (tw *Twelve) ImmutableList(lstHash string, limit pagination.Limit) ([]*ImmutableTx, string, *errors.Error) {
	return tw.queue.ImmutableList(lstHash, limit)
}
