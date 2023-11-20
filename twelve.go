package twelve

import (
	"github.com/hootuu/tome/bk"
	"github.com/hootuu/tome/pr"
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
	chain    bk.Chain
	queue    *Queue
	notifier *Notifier
	peer     *pr.Local
	here     *Peer
	option   *Option
	lock     sync.Mutex
}

func NewTwelve(
	chain bk.Chain,
	twelveNode ITwelveNode,
	option *Option,
) (*Twelve, *errors.Error) {
	if err := chain.Verify(); err != nil {
		return nil, err
	}
	if twelveNode == nil {
		return nil, errors.Verify("require twelve node")
	}
	peer := twelveNode.Peer()
	if peer == nil {
		return nil, errors.Verify("require twelve node peer")
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
		peer:     peer,
		option:   option,
	}
	var err *errors.Error
	tw.queue, err = NewQueue(peer.ID, tw.chain.S())
	if err != nil {
		return nil, err
	}
	tw.notifier.BindListener(tw)
	herePeer, err := peer.Peer()
	if err != nil {
		return nil, err
	}
	tw.here = PeerOf(herePeer)
	return tw, nil
}

func (tw *Twelve) OnRequest(msg *Message) {
	tx, err := tw.queue.Append(msg)
	if err != nil {
		gLogger.Error("twelve.queue.Append(msg) failed",
			zap.Any("hash", msg), zap.Error(err))
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
	requestPayload, err := tx.Message.GetRequestPayload()
	if err != nil {
		return err
	}
	if requestPayload == nil {
		gLogger.Error("invalid message, require Request Payload", zap.Any("tx", tx))
		return errors.Sys("invalid message, require Request Payload")
	}
	prepareMsg, err := NewMessage(PrepareMessage, &ReplyPayload{
		ID: tx.Hash,
	}, tw.here, tw.peer.PRI)
	if err != nil {
		return err
	}
	expect := gExpectFactory.Build(tw.peer.ID, tx.Hash, PrepareMessage, tw.option.Expect)
	go expect.Waiting(func() *errors.Error {
		return tw.notifier.Notify(prepareMsg)
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
	replyPayload, err := tx.Message.GetReplyPayload()
	if err != nil {
		return err
	}
	if replyPayload == nil {
		return errors.Sys("invalid message, require Reply Payload")
	}
	committedMsg, err := NewMessage(CommittedMessage, &ReplyPayload{
		ID: tx.Hash,
	}, tw.here, tw.peer.PRI)
	if err != nil {
		return err
	}
	expect := gExpectFactory.Build(tw.peer.ID, tx.Hash, CommittedMessage, tw.option.Expect)
	go expect.Waiting(func() *errors.Error {
		return tw.notifier.Notify(committedMsg)
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
	replyPayload, err := tx.Message.GetReplyPayload()
	if err != nil {
		return err
	}
	if replyPayload == nil {
		return errors.Sys("invalid message, require Reply Payload")
	}
	confirmedMsg, err := NewMessage(ConfirmedMessage, &ReplyPayload{
		ID: tx.Hash,
	}, tw.here, tw.peer.PRI)
	if err != nil {
		return err
	}
	expect := gExpectFactory.Build(tw.peer.ID, tx.Hash, ConfirmedMessage, tw.option.Expect)
	go expect.Waiting(func() *errors.Error {
		return tw.notifier.Notify(confirmedMsg)
	}, func() {
		_, err := tw.queue.Confirm(tx.Hash)
		if err != nil {
			gLogger.Error("tw.queue.Confirm(tx.Hash) failed", zap.String("hash", tx.Hash))
		}
	})
	return nil
}

func (tw *Twelve) doOnMessage(msg *Message) {
	//if sys.RunMode.IsRd() {
	//	gLogger.Info("twelve.on.message", zap.String("id", msg.ID),
	//		zap.Int("type", int(msg.Type)))
	//}
	replyPayload, err := msg.GetReplyPayload()
	if err != nil {
		gLogger.Error("doOnMessage: msg is invalid",
			zap.Any("msg", msg), zap.Error(err))
		return
	}
	if replyPayload == nil {
		gLogger.Error("doOnMessage: msg is invalid, replyPayload==nil",
			zap.Any("msg", msg))
		return
	}
	hash := replyPayload.ID
	expect, err := gExpectFactory.MustGet(tw.peer.ID, hash, msg.Type)
	if err != nil {
		gLogger.Error("no such expect[PrepareMessage]", zap.String("hash", hash))
		return
	}
	go expect.Reply(msg.Peer.ID)
}

func (tw *Twelve) OnPrepare(msg *Message) {
	tw.doOnMessage(msg)
}

func (tw *Twelve) OnCommitted(msg *Message) {
	tw.doOnMessage(msg)
}

func (tw *Twelve) OnConfirmed(msg *Message) {
	tw.doOnMessage(msg)
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
