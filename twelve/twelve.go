package twelve

import (
	"fmt"
	"github.com/hootuu/tome/bk"
	"github.com/hootuu/tome/pr"
	"github.com/hootuu/utils/errors"
	"sync"
)

type Option struct {
	Expect int `json:"expect"`
}

type Twelve struct {
	chain    bk.Chain
	queue    Queue
	notifier Notifier
	peer     *pr.Local
	option   *Option
	lock     sync.Mutex
}

func NewTwelve(
	chain bk.Chain,
	queue Queue,
	notifier Notifier,
	peer *pr.Local,
	option *Option,
) *Twelve {
	return &Twelve{
		chain:    chain,
		queue:    queue,
		notifier: notifier,
		peer:     peer,
		option:   option,
	}
}

func (twelve *Twelve) Request(msg *Message) *errors.Error {
	j, err := twelve.queue.Append(msg)
	if err != nil {
		return err
	}
	if j.State == Pending {
		return nil
	}

	return nil
}

func (twelve *Twelve) doPrepare(hash string) *errors.Error {
	fmt.Println("do prepare: ", hash)
	j, err := twelve.queue.MustGet(hash)
	if err != nil {
		return err
	}
	requestPayload := j.Message.GetRequestPayload()
	if requestPayload == nil {
		return errors.Sys("invalid message, require Request Payload")
	}
	herePeer, err := twelve.peer.Peer()
	if err != nil {
		return err
	}
	prepareMsg, err := NewMessage(PrepareMessage, &ReplyPayload{
		ID:        j.Message.ID,
		Signature: j.Message.Signature,
	}, PeerOf(herePeer), twelve.peer.PRI)
	if err != nil {
		return err
	}
	expect := NewExpect(j.Hash, twelve.option.Expect)
	go func() {
		expect.Waiting()
		twelve.doCommit(j.Hash)
	}()
	err = twelve.notifier.Notify(prepareMsg)
	if err != nil {
		expect.Cancel()
		return err
	}
	return nil
}

func (twelve *Twelve) OnPrepare(msg *Message) *errors.Error {
	replyPayload := msg.GetReplyPayload()
	if replyPayload == nil {
		return errors.Sys("@OnPrepare invalid message, require Reply Payload")
	}
	return nil
}

func (twelve *Twelve) doCommit(hash string) *errors.Error {
	j, err := twelve.queue.MustGet(hash)
	if err != nil {
		return err
	}
	replyPayload := j.Message.GetReplyPayload()
	if replyPayload == nil {
		return errors.Sys("invalid message, require Reply Payload")
	}
	herePeer, err := twelve.peer.Peer()
	if err != nil {
		return err
	}
	prepareMsg, err := NewMessage(CommittedMessage, &ReplyPayload{
		ID:        j.Message.ID,
		Signature: j.Message.Signature,
	}, PeerOf(herePeer), twelve.peer.PRI)
	if err != nil {
		return err
	}
	expect := NewExpect(j.Hash, twelve.option.Expect)
	go func() {
		expect.Waiting()
		twelve.doConfirm(j.Hash)
	}()
	err = twelve.notifier.Notify(prepareMsg)
	if err != nil {
		expect.Cancel()
		return err
	}
	return nil
}

func (twelve *Twelve) OnCommit(msg *Message) *errors.Error {
	replyPayload := msg.GetReplyPayload()
	if replyPayload == nil {
		return errors.Sys("@OnCommit invalid message, require Reply Payload")
	}
	return nil
}

func (twelve *Twelve) doConfirm(hash string) *errors.Error {
	j, err := twelve.queue.MustGet(hash)
	if err != nil {
		return err
	}
	replyPayload := j.Message.GetReplyPayload()
	if replyPayload == nil {
		return errors.Sys("invalid message, require Reply Payload")
	}
	herePeer, err := twelve.peer.Peer()
	if err != nil {
		return err
	}
	prepareMsg, err := NewMessage(ConfirmedMessage, &ReplyPayload{
		ID:        j.Message.ID,
		Signature: j.Message.Signature,
	}, PeerOf(herePeer), twelve.peer.PRI)
	if err != nil {
		return err
	}
	expect := NewExpect(j.Hash, twelve.option.Expect)
	go func() {
		expect.Waiting()
		twelve.doCommit(j.Hash)
	}()
	err = twelve.notifier.Notify(prepareMsg)
	if err != nil {
		expect.Cancel()
		return err
	}
	return nil
}

func (twelve *Twelve) OnConfirm(msg *Message) *errors.Error {
	replyPayload := msg.GetReplyPayload()
	if replyPayload == nil {
		return errors.Sys("@OnConfirm invalid message, require Reply Payload")
	}
	return nil
}
